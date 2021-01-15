package config

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr"
	logrTesting "github.com/go-logr/logr/testing"
)

func TestGetEnv(t *testing.T) {
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
