package main

import (
	"log/slog"
	"os"
)

func main() {
	cfg, mapping := parseInput(os.Args)

	slog.Info("Starting CEC keyboard handler",
		"cec-cfg", cfg,
		"mappingCount", len(mapping),
	)

	ctx := signalAwareContext()

	handler, err := newHandler(cfg, mapping)
	panicIfErr(err, "Failed to create handler")
	defer func() { panicIfErr(handler.Close(), "Failed to close handler") }()

	slog.Info("Handler created successfully, starting CEC connection")
	handler.Do(ctx)
	slog.Info("CEC connection closed, exiting program")
}
