package config

// Bin a representation on what to download
type Bin struct {
	Cli          string `yaml:"cli"`
	Owner        string `yaml:"owner"`
	Repo         string `yaml:"repo"`
	Match        string `yaml:"match"`
	BaseURL      string `yaml:"baseURL"`
	Download     bool   `yaml:"download"`
	NonGithubURL string `yaml:"nonGithubURL"`
}

// Items config file struct
type Items struct {
	Bins         []Bin  `yaml:"bins"`
	GitHubAPIkey string `yaml:"githubAPIkey"`
	HTTPtimeout  int    `yaml:"httpTimeout"`
	HTTPinsecure bool   `yaml:"httpInsecure"`
	SaveLocation string `yaml:"saveLocation"`
}
