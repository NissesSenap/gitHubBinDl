package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/NissesSenap/gitHubBinDl/pkg/app"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func main() {

	var log logr.Logger

	//context := context.Background()
	zapLog, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)
	ctx := logr.NewContext(context.Background(), log)

	// in the future get from a config file
	// TODO fix from config file
	mygithubToken := ""

	apiKEY := getEnv(ctx, "GITHUBAPIKEY", "")
	if apiKEY != "" {
		mygithubToken = apiKEY
	}

	// creates http client if needed for a redirect

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(5),
	}

	err = app.App(ctx, httpClient, mygithubToken)
	if err != nil {
		fmt.Println(err)
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
