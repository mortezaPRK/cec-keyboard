package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/mortezaPRK/cec-keyboard/cec"
	"github.com/sashko/go-uinput"
)

type cecConfig struct {
	Adapter string
	Name    string
	Type    string
	Mapping map[int][]MappingAction
}

type handler struct {
	k uinput.Keyboard
	m uinput.Mice
	c *cec.Connection

	km map[int][]MappingAction
}

func newHandler(cfg *cecConfig) (*handler, error) {
	slog.Debug("Creating uinput mice")

	mice, err := uinput.CreateMice(0, 0, 0, 0)
	if err != nil {
		return nil, errors.New("failed to create uinput mice: " + err.Error())
	}
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
		k:  keyboard,
		m:  mice,
		c:  cecConn,
		km: cfg.Mapping,
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
		err = errors.Join(err, runWithTimeout(func() { h.c.Destroy() }))
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
	actionToDo, found := h.km[keyPressed.KeyCode]
	if !found {
		slog.Info("Key not found in mapping", "key", keyPressed.KeyCode)
		return
	}

	for _, k := range actionToDo {
		if k.Keyboard != nil {
			switch k.Keyboard.Type {
			case ActionPress:
				panicIfErr(h.k.KeyPress(k.Keyboard.Code), "Failed to send key press", keyPressed.KeyCode)
			case ActionDown:
				panicIfErr(h.k.KeyDown(k.Keyboard.Code), "Failed to send key down", keyPressed.KeyCode)
			case ActionUp:
				panicIfErr(h.k.KeyUp(k.Keyboard.Code), "Failed to send key up", keyPressed.KeyCode)
			}

			continue
		}

		if k.Mouse != nil {
			switch {
			case k.Mouse.LeftClick != nil:
				panicIfErr(h.m.LeftClick(), "Failed to press left button", keyPressed.KeyCode)
			case k.Mouse.RightClick != nil:
				panicIfErr(h.m.RightClick(), "Failed to press right button", keyPressed.KeyCode)
			case k.Mouse.MiddleClick != nil:
				panicIfErr(h.m.MiddleClick(), "Failed to press middle button", keyPressed.KeyCode)
			case k.Mouse.SideClick != nil:
				panicIfErr(h.m.SideClick(), "Failed to press side button", keyPressed.KeyCode)
			case k.Mouse.MoveX != nil:
				panicIfErr(h.m.MoveX(*k.Mouse.MoveX), "Failed to move mice X", *k.Mouse.MoveX)
			case k.Mouse.MoveY != nil:
				panicIfErr(h.m.MoveY(*k.Mouse.MoveY), "Failed to move mice Y", *k.Mouse.MoveY)
			}
		}
	}
}
