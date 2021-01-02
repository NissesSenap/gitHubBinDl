package app

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/NissesSenap/gitHubBinDl/pkg/config"

	"github.com/go-logr/logr"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v33/github"
)

// App start the app
func App(ctx context.Context, httpClient *http.Client, configItem config.Items) error {
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

	pattern := "tkn_0.15.0_Linux_x86_64.tar.gz"

	// TODO currently hardcoded, change to for loop
	err := downloadBin(ctx, client, httpClient, configItem.Bins[0].Owner, configItem.Bins[0].Repo, configItem.Bins[0].Cli, configItem.SaveLocation, pattern)
	if err != nil {
		return err
	}
	return nil
}

func downloadBin(ctx context.Context, client *github.Client, httpClient *http.Client, owner, repo, cliName, saveLocation, pattern string) error {
	log := logr.FromContext(ctx)

	// response gives information about rate limit etc. I assume I will get an error if i go over my rate limit
	// TODO here a log.debug would be nice...
	resp, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return err
	}
	for _, asset := range resp.Assets {
		log.Info(*asset.Name)
		// TODO turn pattern in to a simple regexp
		if *asset.Name == pattern {
			fmt.Println("hit")

			rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, *asset.ID, httpClient)
			if err != nil {
				return err
			}

			// TODO add a if looking for tar.gz or .zip
			err = Untar(saveLocation, cliName, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}

// Untar tar.gz files and put the result in any folder you want this can be something like a /tmp/myNewFolder or /usr/local/bin
func Untar(dst string, cliName string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	cliFile := filepath.Join(dst, cliName)

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

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if err := makeDirectoryIfNotExists(target); err != nil {
				return err
			}

		// if it's a file create it
		case tar.TypeReg:
			// Only write the cliFile
			if target == cliFile {

				/* Since I only untar the cli it self I enforce 0755
				   else use os.FileMode(header.Mode) to get what the filed had when it was tared.
				*/
				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(0755))
				if err != nil {
					return err
				}

				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					return err
				}

				// manually close here after each file operation; defering would cause each file close
				// to wait until all operations have completed.
				f.Close()
			}
		}
	}
}
