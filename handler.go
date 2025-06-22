package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/mortezaPRK/cec-keyboard/cec"
	"github.com/sashko/go-uinput"
)

type action uint8

const (
	Press action = iota
	Down
	Up
)

type keyAction struct {
	code    uint16
	actions action
}
type keyMapping map[int][]keyAction

type cecConfig struct {
	Adapter string
	Name    string
	Type    string
}

type handler struct {
	k uinput.Keyboard
	c *cec.Connection
	m keyMapping
}

func newHandler(cfg cecConfig, mapping keyMapping) (*handler, error) {
	slog.Debug("Creating uinput keyboard")
	keyboard, err := uinput.CreateKeyboard()
	if err != nil {
		return nil, errors.New("failed to create uinput keyboard: " + err.Error())
	}

	slog.Debug("Opening CEC connection", "adapter", cfg.Adapter, "name", cfg.Name, "type", cfg.Type)
	cecConn, err := cec.Open(cfg.Adapter, cfg.Name, cfg.Type)
	if err != nil {
		return nil, errors.New("failed to open CEC connection: " + err.Error())
	}

	slog.Debug("Devices are ready")

	return &handler{
		k: keyboard,
		c: cecConn,
		m: mapping,
	}, nil
}

func (h *handler) Close() (err error) {
	slog.Info("Closing handler resources")
	if h == nil {
		return
	}

	if h.k != nil {
		slog.Info("Closing uinput keyboard")
		err = errors.Join(err, h.k.Close())
		h.k = nil
	}

	if h.c != nil {
		slog.Info("Closing CEC connection")
		err = errors.Join(err, closeCecWithTimeout(h.c))
		h.c = nil
	}

	return
}

func (h *handler) Do(c context.Context) {
	slog.Info("Starting CEC event handler")
	cec.CallbackEvent = h.onCb

	<-c.Done()
	slog.Info("Stopping CEC event handler")

	cec.CallbackEvent = nil
}

func (h *handler) onCb(keyPressed cec.KeyPress) {
	slog.Info("Received CEC key press event", "key", keyPressed.KeyCode, "dur", keyPressed.Duration)
	keysToSend, found := h.m[keyPressed.KeyCode]
	if !found {
		slog.Info("Key not found in mapping", "key", keyPressed.KeyCode)
		return
	}

	for _, k := range keysToSend {
		switch k.actions {
		case Press:
			panicIfErr(h.k.KeyPress(k.code), "Failed to send key press", keyPressed.KeyCode)
		case Down:
			panicIfErr(h.k.KeyDown(k.code), "Failed to send key down", keyPressed.KeyCode)
		case Up:
			panicIfErr(h.k.KeyUp(k.code), "Failed to send key up", keyPressed.KeyCode)
		}
	}
}
