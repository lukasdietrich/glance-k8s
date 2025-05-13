package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/lukasdietrich/glance-k8s/internal/extension"
	"github.com/lukasdietrich/glance-k8s/internal/k8s"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	client, err := k8s.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to cluster: %w", err)
	}

	return http.ListenAndServe(":8080", extension.New(client))
}
