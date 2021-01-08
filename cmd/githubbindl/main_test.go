package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NissesSenap/gitHubBinDl/pkg/config"
	"gopkg.in/yaml.v2"
)

func TestMain(t *testing.T) {

	filename := "../../data.yaml"
	os.Setenv("CONFIGFILE", filename)
	main()

	var item config.Items

	// Read the config file
	source, _ := ioutil.ReadFile(filename)

	// unmarshal the data
	_ = yaml.Unmarshal(source, &item)

	var downloadItems []string
	for i := range item.Bins {
		downloadItems = append(downloadItems, filepath.Join(item.SaveLocation, item.Bins[i].Cli))
	}

	// The /* is due to that the location might be missing a final /, even if it do it won't create any issues
	files, _ := filepath.Glob(filepath.Join(item.SaveLocation + "/*"))

	assert.ElementsMatch(t, downloadItems, files)
}
