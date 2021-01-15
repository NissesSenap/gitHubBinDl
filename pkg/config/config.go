package config

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/NissesSenap/gitHubBinDl/build"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
)

// Bin a representation on what to download
type Bin struct {
	Cli                string   `yaml:"cli"`
	Owner              string   `yaml:"owner"`
	Repo               string   `yaml:"repo"`
	Tag                string   `yaml:"tag"`
	Match              string   `yaml:"match"`
	BaseURL            string   `yaml:"baseURL"`
	Download           bool     `yaml:"download"`
	NonGithubURL       string   `yaml:"nonGithubURL"`
	Backup             bool     `yaml:"backup"`
	CompletionLocation string   `yaml:"completionLocation"`
	CompletionArgs     []string `yaml:"completionArgs"`
}

// Items config file struct
type Items struct {
	Bins         []Bin  `yaml:"bins"`
	GitHubAPIkey string `yaml:"githubAPIkey"`
	HTTPtimeout  int    `yaml:"httpTimeout"`
	HTTPinsecure bool   `yaml:"httpInsecure"`
	SaveLocation string `yaml:"saveLocation"`
}

// ManageConfig read all the user input and returns Items
func ManageConfig(ctx context.Context) Items {
	log := logr.FromContext(ctx)
	var filename string
	configFile := readCli()

	if *configFile == "" {
		filename = getEnv(ctx, "CONFIGFILE", "data.yaml")
	} else {
		filename = *configFile
	}

	log.Info("", "Configfilename:", filename)
	var item Items

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

	return item
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
