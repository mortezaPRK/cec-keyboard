package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/robbiet480/cec"
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

	flag.Parse()

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

// This function solely exists to ensure that a connection is created to CEC without blocking the main goroutine.
//
//	There are two main reasons for this:
//
// 1. The CEC library requires a receiver for the event channel to avoid blocking.
// 2. Timeouts are necessary to prevent the program from hanging indefinitely if the CEC connection cannot be established.
func createCecWithTimeout(
	readyForEvents <-chan struct{},
	cfg cecConfig,
) *cec.Connection {
	var mu sync.Mutex

	// Without this, opening a connection will stuck since there is no receiver for the event channel.
	mu.Lock()
	go func() {
		mu.Unlock()
		slog.Debug("Waiting for CEC events before processing")
		timeout := time.After(20 * time.Second)
		for {
			select {
			case <-readyForEvents:
				slog.Debug("Ready for CEC events")
				return
			case <-cec.CallbackEvents:
				slog.Debug("Received CEC event, but not ready for events yet")
				// Consume events to prevent blocking.
				// This is a workaround to avoid blocking the CEC event channel.
			case <-timeout:
				slog.Warn("Timeout waiting for CEC events, no receiver ready")
				panic("Timeout waiting for CEC events, no receiver ready")
			}
		}
	}()

	// Make sure the above goroutine has started before proceeding to open the CEC connection.
	mu.Lock()
	defer mu.Unlock()
	time.Sleep(5 * time.Second)

	slog.Debug("Opening CEC connection", "adapter", cfg.Adapter, "name", cfg.Name, "type", cfg.Type)
	cecConn, err := cec.Open(cfg.Adapter, cfg.Name, cfg.Type)
	panicIfErr(err, "failed to open CEC connection", "adapter", cfg.Adapter, "name", cfg.Name, "type", cfg.Type)

	return cecConn
}
