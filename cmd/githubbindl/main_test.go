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

	// Create a list of all the items from the configFile
	var downloadItems []string
	for i := range item.Bins {
		downloadItems = append(downloadItems, filepath.Join(item.SaveLocation, item.Bins[i].Cli))
	}

	// Create a list of files that Is possible to stat
	var statFiles []string
	for _, fileLocation := range downloadItems {
		srcStat, err := os.Stat(fileLocation)
		if err != nil {
			t.Errorf("One or more files is missing %v", err)
		}
		if srcStat.Mode().IsRegular() {
			statFiles = append(statFiles, fileLocation)
		}
	}

	// compare what should exists with what actually is saved on disk
	assert.ElementsMatch(t, downloadItems, statFiles)
}
