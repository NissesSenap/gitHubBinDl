package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/NissesSenap/gitHubBinDl/pkg/app"
	"github.com/NissesSenap/gitHubBinDl/pkg/config"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func main() {

	var log logr.Logger

	//context := context.Background()
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)
	ctx := logr.NewContext(context.Background(), log)

	// TODO handle this with CLI input instead
	filename := getEnv(ctx, "CONFIGFILE", "data.yaml")

	var item config.Items

	// Read the config file
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error(err, "Unable to read configfile")
		os.Exit(1)
	}

	// unmarshal the data
	err = yaml.Unmarshal(source, &item)
	if err != nil {
		log.Error(err, "Unable to read values from configfile")
		os.Exit(1)

	}

	// TODO log.Debug("config" item) don't work need some basic help here
	//fmt.Printf("config: %v", item)

	apiKEY := getEnv(ctx, "GITHUBAPIKEY", "")
	if apiKEY != "" {
		item.GitHubAPIkey = apiKEY
	}

	// creates http client if needed for a redirect

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: item.HTTPinsecure},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(5),
	}

	err = app.App(ctx, httpClient, &item)
	if err != nil {
		log.Error(err, "Unable to download bins")
		os.Exit(1)
	}
}

// getEnv get key environment variable if exist otherwise return defalutValue
func getEnv(ctx context.Context, key, defaultValue string) string {
	log := logr.FromContext(ctx)

	if value, exists := os.LookupEnv(key); exists {
		log.Info(value)

		return value
	}
	return defaultValue
}
