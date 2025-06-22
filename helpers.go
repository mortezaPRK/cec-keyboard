package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mortezaPRK/cec-keyboard/cec"
)

func closeCecWithTimeout(c *cec.Connection) error {
	if c == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		c.Destroy()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("CEC connection closed successfully")
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for CEC connection to close")
	}
}

func parseInput(args []string) (cecConfig, keyMapping) {
	var (
		adapter string
		name    string
		cecType string
	)
	mapping := make(keyMapping)

	flag.StringVar(&adapter, "adapter", "", "CEC adapter name (optional)")
	flag.StringVar(&name, "name", "", "CEC device name (optional)")
	flag.StringVar(&cecType, "type", "recording", "CEC device type (optional, default: recording), one of tv,recording,tuner,playback,audio")
	flag.Func("log-level", "Set log level (debug, info, warn, error)", func(s string) error {
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
	})
	flag.Func("mapping", "key mapping: cecCode=p:keyCode1,u:keyCode2,h:keyCode3,d:keyCode4", func(s string) error {
		k, v := parseMapping(s)

		if _, exists := mapping[k]; exists {
			return fmt.Errorf("duplicate mapping for key %d", k)
		}

		mapping[k] = v

		return nil
	})

	flag.CommandLine.Parse(args[1:])

	return cecConfig{
		Adapter: adapter,
		Name:    name,
		Type:    cecType,
	}, mapping
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
		slog.Error("Force exiting due to second termination signal")
		os.Exit(1)
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
