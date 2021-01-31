package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NissesSenap/gitHubBinDl/build"
	"github.com/NissesSenap/gitHubBinDl/pkg/util"
	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	Bins                []Bin    `yaml:"bins"`
	GitHubAPIkey        string   `yaml:"githubAPIkey"`
	HTTPtimeout         int      `yaml:"httpTimeout"`
	HTTPinsecure        bool     `yaml:"httpInsecure"`
	SaveLocation        string   `yaml:"saveLocation"`
	MaxFileSize         int64    `yaml:"maxFileSize"`
	NotOkCompletionArgs []string `yaml:"notOkCompletionArgs"`
}

// default Keys & values for global values lik saveLocation & HttpTimeout, notice that only the keys are Global
const (
	DefaultGITHUBAPIKEYKey   = "gitHubAPIkey"
	defaultGITHUBAPIKEYValue = ""

	DefaultConfigFileKey   = "configfile"
	defaultConfigFileValue = "data.yaml"

	DefaultHTTPtimeoutkey   = "httpTimeout"
	defaultHTTPtimeoutValue = 5

	DefaultHTTPinsecureKey   = "httpInsecure"
	defaultHTTPinsecureValue = false

	DefaultSaveLocationKey = "saveLocation"
	// defaultSaveLocationValue is defined in ManageConfig()

	DefaultMaxFileSizeKey   = "maxFileSize"
	defaultMaxFileSizeValue = int64(104857600) //1024*1024*100 aka 100 Mb

	DefaultNotOkCompletionArgsKey = "notOkCompletionArgs"
	//defaultNotOkCompletionArgsValue is defined in ManageConfig()
)

// ManageConfig read all the user input and returns Items
func ManageConfig(ctx context.Context) (Items, error) {
	log := logr.FromContext(ctx)
	var item Items

	// default value for saveLocation is user homedir + gitGubBinDL_<todays-date> example: 2021-01-24
	homedir, err := os.UserHomeDir()
	if err != nil {
		return item, err
	}
	defaultSaveLocationValue := filepath.Join(homedir, "gitGubBinDL"+"_"+time.Now().Local().Format(util.DateFormat))
	var defaultNotOkCompletionArgsValue = []string{"sudo", "rm", "ln", "sed", "awk", "|", "&"}

	//var filename string
	err = readCli()
	if err != nil {
		return item, err
	}

	// set default value
	viper.SetDefault(DefaultGITHUBAPIKEYKey, defaultGITHUBAPIKEYValue)
	viper.SetDefault(DefaultConfigFileKey, defaultConfigFileValue)
	viper.SetDefault(DefaultHTTPtimeoutkey, defaultHTTPtimeoutValue)
	viper.SetDefault(DefaultHTTPinsecureKey, defaultHTTPinsecureValue)
	viper.SetDefault(DefaultSaveLocationKey, defaultSaveLocationValue)
	viper.SetDefault(DefaultMaxFileSizeKey, defaultMaxFileSizeValue)
	viper.SetDefault(DefaultNotOkCompletionArgsKey, defaultNotOkCompletionArgsValue)

	// Run initial to look if env FILE exists to manage the config
	viper.AutomaticEnv()

	// Grab the value from viper that we got from readCli()
	configValue := viper.GetString(DefaultConfigFileKey)
	log.Info(configValue)

	dir, file := filepath.Split(configValue)
	log.Info("Config:", dir, file)
	viper.SetConfigName(file)
	viper.SetConfigType("yaml")

	viper.AddConfigPath(dir)
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		return item, err
	}
	err = viper.Unmarshal(&item)
	if err != nil {
		return item, err
	}

	// run again to get the values from the config file like GITHUBAPIKEY
	viper.AutomaticEnv()
	return item, nil
}

func readCli() error {

	help := pflag.BoolP("help", "h", false, "prints the help output.")
	_ = pflag.StringP(DefaultConfigFileKey, "c", "", "Configfile to read data from, default data.yaml")
	version := pflag.BoolP("version", "v", false, "print application version.")
	//pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}

	if *help {
		pflag.PrintDefaults()
		os.Exit(0)
	}

	if *version {
		fmt.Printf("githubBinDl Version: %s, BuildDate: %s", build.Version, build.BuildDate)
		os.Exit(0)
	}
	return nil
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
