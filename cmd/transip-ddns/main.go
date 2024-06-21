package main

import (
	"github.com/hsmade/transip-ddns/pkg/ddns"
	"github.com/transip/gotransip/v6"
	"log/slog"
	"os"
	"strings"
)

func main() {
	slog.Info("starting")

	if os.Getenv("VERBOSE") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	accountName := os.Getenv("ACCOUNT_NAME")
	if accountName == "" {
		slog.Error("ACCOUNT_NAME environment variable not set")
		os.Exit(-1)
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		slog.Error("TOKEN environment variable not set")
		os.Exit(-1)
	}

	domainNames := strings.Split(os.Getenv("DOMAIN_NAMES"), ",")
	if len(domainNames) == 0 {
		slog.Error("DOMAIN_NAMES environment variable not set")
		os.Exit(-1)
	}

	client, err := gotransip.NewClient(gotransip.ClientConfiguration{
		AccountName: accountName,
		Token:       token,
	})
	if err != nil {
		slog.Error("Error creating transip client:", err)
		os.Exit(-1)
	}

	updater := ddns.DDNS{
		DomainNames: domainNames,
		Client:      client,
	}

	err = updater.Update()
	if err != nil {
		slog.Error("Error updating domain names:", "error", err)
	}
	slog.Info("done")
}
