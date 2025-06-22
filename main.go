package main

import (
	"flag"
	"log/slog"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.yaml", "Path to the configuration file")
	flag.Func("log-level", "Set log level (debug, info, warn, error)", setLogger)
	flag.Parse()

	config, err := LoadConfig(cfgPath)
	panicIfErr(err, "Failed to load configuration")

	slog.Info("Starting CEC keyboard handler")
	ctx := signalAwareContext()

	handler, err := newHandler(config)
	panicIfErr(err, "Failed to create handler")
	defer func() { panicIfErr(handler.Close(), "Failed to close handler") }()

	slog.Info("Handler created successfully, starting CEC connection")
	handler.Do(ctx)
	slog.Info("CEC connection closed, exiting program")
}
