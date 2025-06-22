package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func runWithTimeout(fn func()) error {
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("Function completed successfully")
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for function to complete")
	}
}

func signalAwareContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received termination signal, shutting down gracefully")
		cancel()

		anotherSigChan := make(chan os.Signal, 1)
		signal.Notify(anotherSigChan, syscall.SIGINT, syscall.SIGTERM)

		<-anotherSigChan
		panic("Force exit due to second termination signal")
	}()

	return ctx
}

func panicIfErr(err error, msg string, args ...any) {
	panicIf(err != nil, msg+": %v", append(args, err)...)
}

func panicIf(condition bool, msg string, args ...any) {
	if condition {
		panic(fmt.Sprintf(msg, args...))
	}
}

func setLogger(s string) error {
	switch strings.ToLower(s) {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "info":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "warn", "warning":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		return fmt.Errorf("invalid log level: %s. Use debug, info, warn, or error", s)
	}

	return nil
}
