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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/NissesSenap/gitHubBinDl/pkg/config"

	"github.com/go-logr/logr"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v33/github"
)

const zipExtension = ".zip"
const gzExtension = ".gz"
const exeExtension = ".exe"
const dateFormat = "2006-01-02"

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

	var wg sync.WaitGroup
	channel := make(chan error, len(configItem.Bins))

	for i := range configItem.Bins {
		// TODO check configItem.Bins[i].Download == false and create a report function that only is called.
		wg.Add(1)
		go downloadBin(ctx, &wg, channel, client, httpClient, configItem.Bins[i], configItem.HTTPtimeout, configItem.SaveLocation)
	}

	// Blocking, waiting for the wg to finish
	wg.Wait()

	// only check for errors, if no error close the channel and return nil
	select {
	case err := <-channel:
		if err != nil {
			close(channel)
			return err
		}
	default:
	}

	close(channel)
	return nil

}

func downloadBin(ctx context.Context, wg *sync.WaitGroup, channel chan error, client *github.Client, httpClient *http.Client, binConfig config.Bin, httpTimeout int, saveLocation string) {
	defer wg.Done()

	log := logr.FromContext(ctx)

	log.Info(binConfig.NonGithubURL)
	if binConfig.NonGithubURL != "" {
		// Instead of using httpClient.Timeout I use a ctx with Deadline.
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Duration(httpTimeout)*time.Second))
		defer cancel()

		req, err := http.NewRequest(http.MethodGet, binConfig.NonGithubURL, nil)
		if err != nil {
			channel <- err
			return
		}
		req = req.WithContext(ctx)
		resp, err := httpClient.Do(req)
		if err != nil {
			channel <- err
			return
		}
		defer resp.Body.Close()

		err = pickExtension(ctx, resp.Body, binConfig.Cli, saveLocation, binConfig.NonGithubURL, binConfig.Backup)
		if err != nil {
			channel <- err
			return
		}

		if binConfig.CompletionLocation != "" {
			err := saveCompletion(ctx, saveLocation, binConfig.Cli, binConfig.CompletionLocation, binConfig.CompletionArgs)
			if err != nil {
				channel <- err
				return
			}
		}
		// Generate the completion file
		return
	}

	var resp *github.RepositoryRelease
	var er error

	// check if BaseURL is empty, if not it will use that when talking to the github api
	if binConfig.BaseURL != "" {
		githubURL, err := url.Parse(binConfig.BaseURL)
		if err != nil {
			channel <- er
			return
		}
		client.BaseURL = githubURL
	}

	// If tag is empty use GetReleaseByTag
	if binConfig.Tag != "" {
		// response gives information about rate limit etc. I assume I will get an error if i go over my rate limit
		// TODO here a log.debug would be nice...
		resp, _, er = client.Repositories.GetReleaseByTag(ctx, binConfig.Owner, binConfig.Repo, binConfig.Tag)
		if er != nil {
			channel <- er
			return
		}

	} else {
		resp, _, er = client.Repositories.GetLatestRelease(ctx, binConfig.Owner, binConfig.Repo)
		if er != nil {
			channel <- er
			return
		}
	}

	for _, asset := range resp.Assets {
		log.Info(*asset.Name)
		lowerAssetName := strings.ToLower(*asset.Name)
		patternMatched, err := regexp.MatchString(strings.ToLower(binConfig.Match), lowerAssetName)
		if err != nil {
			channel <- err
			return
		}
		if patternMatched {
			rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, binConfig.Owner, binConfig.Repo, *asset.ID, httpClient)
			if err != nil {
				channel <- err
				return
			}
			err = pickExtension(ctx, rc, binConfig.Cli, saveLocation, lowerAssetName, binConfig.Backup)
			if err != nil {
				channel <- err
				return
			}

			// Generate the completion file
			if binConfig.CompletionLocation != "" {
				err := saveCompletion(ctx, saveLocation, binConfig.Cli, binConfig.CompletionLocation, binConfig.CompletionArgs)
				if err != nil {
					channel <- err
					return
				}
			}

			return
		}
	}

	// normally return earlier, should only come here if we fail to find the bin
	channel <- errors.New("Unable to find match")
}

// copyOldCli copies the current cli to the same location but with addition of _2006-01-02
func copyOldCli(cliName, saveLocation string) error {
	target := filepath.Join(saveLocation, cliName)

	dst := target + "_" + time.Now().Local().Format(dateFormat)

	srcStat, err := os.Stat(target)
	if err != nil {
		return err
	}
	if !srcStat.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", srcStat.Name(), srcStat.Mode().String())
	}
	dstStat, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !(dstStat.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dstStat.Name(), dstStat.Mode().String())
		}
		// I don't understand how SameFile works, even though the files are copies of eachother they still return false.
		// Well in some special case it could speed it up by not having to copy the data.
		if os.SameFile(srcStat, dstStat) {
			return nil
		}
	}
	err = copyFileContents(target, dst, srcStat.Mode())
	if err != nil {
		return err
	}
	return nil
}

// copyFileContents actually performs the copy and uses the existing FileMode to set the old one
func copyFileContents(target, dst string, srcStat os.FileMode) error {
	in, err := os.Open(target)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	err = os.Chmod(dst, srcStat)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return nil
}

func pickExtension(ctx context.Context, respBody io.ReadCloser, cliName, saveLocation, downloadURL string, backup bool) error {

	log := logr.FromContext(ctx)

	if backup {
		err := copyOldCli(cliName, saveLocation)
		if err != nil {
			// The application will continue and instead overwrite the existing cliName
			log.Info("msg", "Unable to save a old version of cli ", err)
		}
	}
	switch filepath.Ext(downloadURL) {
	case gzExtension:
		err := untarGZ(ctx, saveLocation, cliName, respBody)
		if err != nil {
			return err
		}
		return nil
	case zipExtension:
		zipRespBody, err := ioutil.ReadAll(respBody)
		if err != nil {
			return err
		}
		err = unZIP(ctx, saveLocation, cliName, zipRespBody)
		if err != nil {
			return err
		}
		return nil
	case "", exeExtension:
		err := saveFile(ctx, saveLocation, cliName, respBody)
		if err != nil {
			return err
		}
	default:
		return errors.New("The file extenssion is not supported")
	}

	return nil
}

//saveFile used if the file have no extension
func saveFile(ctx context.Context, dst, cliName string, rc io.Reader) error {
	log := logr.FromContext(ctx)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(rc)
	if err != nil {
		return err
	}

	target := filepath.Join(dst, cliName)
	f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0755))
	if err != nil {
		return err
	}

	log.Info("Downloading", "target", target)
	_, err = buf.WriteTo(f)
	if err != nil {
		return err
	}
	//CLOSE THE FILE
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func saveCompletion(ctx context.Context, cliLocation, cliName, completionLocation string, completionCommand []string) error {
	log := logr.FromContext(ctx)
	command := exec.Command(filepath.Join(cliLocation, cliName))

	// Instead of using a for loop with append you can use ... to unpack the list S1011
	command.Args = append(command.Args, completionCommand...)

	var out bytes.Buffer

	// set the output to our variable
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return err
	}
	log.Info("Managed to run completion command")

	f, err := os.OpenFile(completionLocation, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0755))
	if err != nil {
		return err
	}

	_, err = out.WriteTo(f)
	if err != nil {
		return err
	}
	//CLOSE THE FILE
	if err := f.Close(); err != nil {
		return err
	}
	return nil

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
