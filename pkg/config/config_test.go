package config

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr"
	logrTesting "github.com/go-logr/logr/testing"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestManageConfig(t *testing.T) {
	IndexOsEnv := []struct {
		setEnv    bool
		key       string
		value     string
		expectOut string
	}{
		{setEnv: false, key: "GITHUBAPIKEY", value: "", expectOut: ""},
		{setEnv: true, key: "GITHUBAPIKEY", value: "myAPIKEY", expectOut: "myAPIKEY"},
	}

	ctx := logr.NewContext(context.Background(), logrTesting.NullLogger{})

	for _, tests := range IndexOsEnv {

		err := os.Unsetenv(tests.key)
		if err != nil {
			t.Errorf("Unable to unset environment variable")
		}

		// viper.Reset() deletes all viper settings, pflag removes any pflag config that is currently stored
		viper.Reset()
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError) //flags are now reset
		// To make the test work, add configPath to data.yaml after the reset
		viper.AddConfigPath("../../")

		if tests.setEnv {
			os.Setenv(tests.key, tests.value)
		}
		_, err = ManageConfig(ctx)
		if err != nil {
			t.Errorf("Unable to read ManageConfig, err: %v", err)
		}

		if viper.GetString(DefaultGITHUBAPIKEYKey) != tests.expectOut {
			t.Errorf("The expected value %v from getEnv, didn't match the actual output: %v ", tests.expectOut, viper.GetString(DefaultGITHUBAPIKEYKey))
		}
	}
}

func TestFuGetEnv(t *testing.T) {
	IndexOsEnv := []struct {
		theDefault string
		setEnv     bool
		key        string
		value      string
		expectOut  string
	}{
		{theDefault: "default1", setEnv: false, key: "test1", value: "", expectOut: "default1"},
		{theDefault: "default2", setEnv: true, key: "test2", value: "", expectOut: ""},
		{theDefault: "default3", setEnv: true, key: "test3", value: "value3", expectOut: "value3"},
	}

	ctx := logr.NewContext(context.Background(), logrTesting.NullLogger{})

	for _, tests := range IndexOsEnv {
		if tests.setEnv {
			os.Setenv(tests.key, tests.value)
		}
		actualOutput := getEnv(ctx, tests.key, tests.theDefault)

		if actualOutput != tests.expectOut {
			t.Errorf("The expected value %v from getEnv, didn't match the actual output: %v ", tests.expectOut, actualOutput)
		}
	}
}
