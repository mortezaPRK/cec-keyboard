package cec

// #include <libcec/cecc.h>
import "C"

import (
	"unsafe"
)

type KeyPress struct {
	KeyCode  int
	Duration int
}

type CallbackFn func(KeyPress)

var CallbackEvent CallbackFn

//export keyPressCallback
func keyPressCallback(_ unsafe.Pointer, keyPress *C.cec_keypress) C.uint8_t {
	if CallbackEvent != nil && keyPress != nil && keyPress.duration > 0 {
		CallbackEvent(KeyPress{
			KeyCode:  int(keyPress.keycode),
			Duration: int(keyPress.duration),
		})
	}
	return 1
}
