package app

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NissesSenap/gitHubBinDl/pkg/config"

	"github.com/go-logr/logr"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v33/github"
)

const zipExtension = ".zip"
const gzExtension = ".gz"

// App start the app
func App(ctx context.Context, httpClient *http.Client, configItem *config.Items) error {
	// TODO find a way to use configItem.Bins[0].BaseURL to download files from custom github endpoints
	client := github.NewClient(nil)

	// If no githuBAPIToken is specified the application runs without it
	if configItem.GitHubAPIkey != "" {
		tokenService := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: configItem.GitHubAPIkey},
		)
		tokenClient := oauth2.NewClient(ctx, tokenService)

		client = github.NewClient(tokenClient)
	}

	// Create the download folder if needed
	if err := makeDirectoryIfNotExists(configItem.SaveLocation); err != nil {
		return err
	}

	for i := range configItem.Bins {
		// TODO check configItem.Bins[i].Download == false and create a report function that only is called.
		err := downloadBin(ctx, client, httpClient, configItem.Bins[i].Owner, configItem.Bins[i].Repo, configItem.Bins[i].Cli, configItem.SaveLocation, configItem.Bins[i].Match, configItem.Bins[i].NonGithubURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadBin(ctx context.Context, client *github.Client, httpClient *http.Client, owner, repo, cliName, saveLocation, pattern, nonGithubURL string) error {
	log := logr.FromContext(ctx)

	log.Info(nonGithubURL)
	if nonGithubURL != "" {
		resp, err := httpClient.Get(nonGithubURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if filepath.Ext(nonGithubURL) == gzExtension {
			err = untarGZ(ctx, saveLocation, cliName, resp.Body)
			if err != nil {
				return err
			}
			return nil
		}

		if filepath.Ext(nonGithubURL) == zipExtension {
			zipRespBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			err = unZIP(ctx, saveLocation, cliName, zipRespBody)
			if err != nil {
				return err
			}
			return nil
		}

		return nil
	}

	// response gives information about rate limit etc. I assume I will get an error if i go over my rate limit
	// TODO here a log.debug would be nice...
	resp, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return err
	}

	for _, asset := range resp.Assets {
		log.Info(*asset.Name)
		lowerAssetName := strings.ToLower(*asset.Name)
		patternMatched, err := regexp.MatchString(strings.ToLower(pattern), lowerAssetName)
		if err != nil {
			return err
		}
		if patternMatched {
			rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, *asset.ID, httpClient)
			if err != nil {
				return err
			}

			if filepath.Ext(lowerAssetName) == gzExtension {
				err = untarGZ(ctx, saveLocation, cliName, rc)
				if err != nil {
					return err
				}
			}

			if filepath.Ext(lowerAssetName) == zipExtension {
				zipRespBody, err := ioutil.ReadAll(rc)
				if err != nil {
					return err
				}
				err = unZIP(ctx, saveLocation, cliName, zipRespBody)
				if err != nil {
					return err
				}
			}

			// return directly when we get a match, no need to keep on running the loop
			return nil
		}
	}

	// normally return earlier, should only come here if we fail to find the bin
	return errors.New("Unable to find match")
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}

// untarGZ tar.gz files and put the result in any folder you want
func untarGZ(ctx context.Context, dst, cliName string, r io.Reader) error {
	log := logr.FromContext(ctx)

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		/* HELM is a pain, the bin file is inside a folder.
		First split the target and now join it with the base dir where we want to save the file.
		*/
		// the target location + header.Name - it's sub folder
		_, cleanHeader := filepath.Split(header.Name)
		target := filepath.Join(dst, cleanHeader)

		// check the file type
		switch header.Typeflag {

		// I will never create a folder in this app when unpacking files...
		// Save for reference
		/*
			case tar.TypeDir:
				if err := makeDirectoryIfNotExists(target); err != nil {
					return err
				}
		*/

		// if it's a file create it
		case tar.TypeReg:
			// Only write the cliFile
			if cleanHeader == cliName {

				// TODO change to some debug...
				log.Info(cleanHeader)
				/* Since I only untar the cli it self I enforce 0755
				   else use os.FileMode(header.Mode) to get what the filed had when it was tared.
				*/
				file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(0755))
				if err != nil {
					return err
				}

				// copy over contents
				if _, err := io.Copy(file, tr); err != nil {
					return err
				}

				// manually close here after each file operation; defering would cause each file close
				// to wait until all operations have completed.
				file.Close()
			}
		}
	}
}

// unZIP unzip files and put the result in any folder you want
func unZIP(ctx context.Context, dst, cliName string, r []byte) error {
	log := logr.FromContext(ctx)

	zipReader, err := zip.NewReader(bytes.NewReader(r), int64(len(r)))
	if err != nil {
		return err
	}

	log.Info("we are in zip")

	for _, f := range zipReader.File {
		// the target location + header.Name - it's sub folder
		_, cleanHeader := filepath.Split(f.Name)
		target := filepath.Join(dst, cleanHeader)

		if cleanHeader == cliName {
			// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
			if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
				return fmt.Errorf("%s: illegal file path", target)
			}

			outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			rc, err := f.Open()
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()
			rc.Close()

			if err != nil {
				return err
			}
		}
	}
	return nil
}
