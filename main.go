package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/google/go-github/v33/github"
)

func main() {

	var log logr.Logger

	//context := context.Background()
	zapLog, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)
	ctx := logr.NewContext(context.Background(), log)

	// creates http client if needed for a redirect
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(5),
	}

	// so it begins
	owner := "tektoncd"
	repo := "cli"
	cliName := "tkn"
	tmpPath := "/tmp/tkn-201230"
	//fulPathTarGz := tmpPath + "/" + cliName + ".tar.gz"

	client := github.NewClient(nil)
	// response gives information about my rate limit etc, nice to have but no need to do stuff. I assume I will get an error if i go over my rate limit
	resp, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		fmt.Println("shiiiiit")
	}
	pattern := "tkn_0.15.0_Linux_x86_64.tar.gz"
	for _, asset := range resp.Assets {
		log.Info(*asset.Name)
		if *asset.Name == pattern {
			fmt.Println("hit")

			// Create the tmp folder
			if err := makeDirectoryIfNotExists(tmpPath); err != nil {
				log.Info("unable to stat folder")
			}

			rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, *asset.ID, httpClient)
			if err != nil {
				log.Info("unable to download stuff")
				return
			}

			/*
				buf := new(bytes.Buffer)
				buf.ReadFrom(rc)

				// notice the tar.gz
				f, err := os.OpenFile(fulPathTarGz, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
				if err != nil {
					log.Error(err, "shiiit")
				}

				// write the file
				buf.WriteTo(f)

				//CLOSE THE FILE
				if err := f.Close(); err != nil {
					log.Info("unable to write file")
					os.Exit(1)
				}
			*/
			err = Untar(tmpPath, cliName, rc)
			if err != nil {
				log.Error(err, "unable to untar files")
			}
		}
		/*
			untar
			chmod 755
			move bin to correct area
			remove tmp/asset folder

		*/
	}

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

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			// Only write the cliFile
			if target == cliFile {

				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
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
