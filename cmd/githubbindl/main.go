package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/NissesSenap/gitHubBinDl/pkg/app"
	"github.com/NissesSenap/gitHubBinDl/pkg/config"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
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

	item := config.ManageConfig(ctx)
	// creates http client if needed for a redirect

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: item.HTTPinsecure},
	}

	// TODO set 5 as the default timeout value, ignore if HTTPtimeout = 0
	httpClient := &http.Client{
		Transport: tr,
	}

	err = app.App(ctx, httpClient, &item)
	if err != nil {
		log.Error(err, "Unable to download bins")
		os.Exit(1)
	}
}
