package main

import (
	"context"
	"log/slog"
	"os"
	"path"

	"github.com/samber/lo"
)

func init() {
	var logLv slog.Level
	if err := logLv.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))); err != nil {
		logLv = slog.LevelWarn
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: logLv == slog.LevelDebug || logLv == slog.LevelError,
		Level:     logLv,
	})))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	arg0, _ := lo.First(os.Args)
	if name := path.Base(arg0); name != "" {
		app.Name = name
	}
	if err := app.Run(ctx, os.Args); err != nil {
		slog.ErrorContext(ctx, "app run finished with error", "err", err.Error())
		os.Exit(1)
	}
}
