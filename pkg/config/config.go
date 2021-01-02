package config

// Bin a representation on what to download
type Bin struct {
	Cli          string // Name of the cli
	Owner        string
	Repo         string
	Match        string // What we should look for, for example Linux_x86_64
	BaseURL      string // default github.com, but can be pointed to a local github enterprise instance
	Download     bool   // If we should download the bin, else it will just write a log default is set to true
	nonGithubURL string // If used all other values except Cli will be ignored and we will just download the file and unpack it
}

// Items config file struct
type Items struct {
	Bins         []Bin  `yaml:"bins"`
	GitHubAPIkey string `yaml:"githubAPIkey"`
	HTTPtimeout  int    `yaml:"httpTimeout"`
	HTTPinsecure bool   `yaml:"httpInsecure"`
	SaveLocation string `yaml:"saveLocation"`
}
