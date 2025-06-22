package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseMapping(inputStr string) (int, []keyAction) {
	// inputStr is expected to be in the format "keyCode=p:code,u:code...."

	parts := strings.Split(strings.TrimSpace(inputStr), "=")
	panicIf(len(parts) != 2, "Invalid mapping format (key=actions): %s", inputStr)

	cecKeyCode, err := strconv.Atoi(parts[0])
	panicIfErr(err, fmt.Sprintf("Invalid CEC key code in mapping: %s", inputStr))

	actions := make([]keyAction, 0)

	for actionStr := range strings.SplitSeq(parts[1], ",") {
		actionParts := strings.Split(actionStr, ":")
		panicIf(len(actionParts) != 2, "Invalid action format in mapping: %s", actionStr)

		uinputKeyCode, err := strconv.ParseUint(actionParts[1], 10, 16)
		panicIfErr(err, fmt.Sprintf("Invalid uinput key code in mapping: %s", actionStr))

		action := keyAction{
			code: uint16(uinputKeyCode),
		}

		switch actionParts[0] {
		case "p":
			action.actions = Press
		case "u":
			action.actions = Up
		case "d":
			action.actions = Down
		case "h":
			// Hold is down and up (stacked)
			action.actions = Down

			defer func(code uint16) {
				action := keyAction{
					code:    code,
					actions: Up,
				}

				actions = append(actions, action)
			}(action.code)
		default:
			panicIf(true, "Invalid action type in mapping: %s", actionParts[0])
		}

		actions = append(actions, action)
	}

	panicIf(len(actions) == 0, "No actions defined for CEC key code %d in mapping: %s", cecKeyCode, inputStr)

	return cecKeyCode, actions
}
