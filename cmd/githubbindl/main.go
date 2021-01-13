package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NissesSenap/gitHubBinDl/build"
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
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)

	// Create SIGTERM channel
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// blocking, waiting for a SIGTERM
		<-done
		// cancel ctx
		cancel()
	}()

	var filename string
	configFile := readCli()

	if *configFile == "" {
		filename = getEnv(ctx, "CONFIGFILE", "data.yaml")
	} else {
		filename = *configFile
	}

	log.Info("", "Configfilename:", filename)
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

	// TODO set 5 as the default timeout value, ignore if HTTPtimeout = 0
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(item.HTTPtimeout),
	}

	err = app.App(ctx, httpClient, &item)
	if err != nil {
		log.Error(err, "Unable to download bins")
		os.Exit(1)
	}
}

func readCli() *string {
	help := flag.Bool("help", false, "prints the help output.")
	configfile := flag.String("f", "", "Configfile to read data from, default data.yaml")
	version := flag.Bool("version", false, "print application version.")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *version {
		fmt.Printf("githubBinDl Version: %s, BuildDate: %s", build.Version, build.BuildDate)
		os.Exit(0)
	}

	return configfile
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
