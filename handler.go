package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/sashko/go-uinput"

	"github.com/robbiet480/cec"
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

	readyForEvents chan struct{}
}

func newHandler(cfg cecConfig, mapping keyMapping) *handler {
	h := &handler{m: mapping, readyForEvents: make(chan struct{})}

	slog.Debug("Creating uinput keyboard")
	keyboard, err := uinput.CreateKeyboard()
	panicIfErr(err, "Failed to create uinput keyboard")

	h.k = keyboard
	h.c = createCecWithTimeout(h.readyForEvents, cfg)

	slog.Debug("Devices are ready")

	return h
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
	slog.Debug("Starting Loop, closing the dummy channel receiver")
	close(h.readyForEvents)
	slog.Debug("Waiting for CEC events")

	for {
		select {
		case <-c.Done():
			slog.Info("Context done, exiting handler loop")
			return
		case e := <-cec.CallbackEvents:
			slog.Debug("Received CEC event", "event", e)
			h.onCb(e)
		}
	}
}

func (h *handler) onCb(e any) {
	keyPressed, ok := e.(cec.KeyPress)
	if !ok {
		slog.Debug("Received non-KeyPress event", "event", e)
		return
	}

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
