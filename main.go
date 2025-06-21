package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sashko/go-uinput"

	"github.com/robbiet480/cec"
)

func main() {
	// Three sub commands:
	// 1. cec: check if CEC is available
	// 2. uinput: check if uinput is available
	// 3. just print HI

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "cec":
		if len(os.Args) != 4 {
			fmt.Println("Usage: go run main.go cec <adapter> <device>")
			return
		}
		checkCecNew(os.Args[2], os.Args[3])
	case "uinput":
		checkUinput()
	case "hi":
		fmt.Println("HI")
	default:
		fmt.Println("Unknown command:", os.Args[1])
		printHelp()
		return
	}
}

func checkCecNew(adapter, device string) {
	var closeChan = make(chan struct{})

	defer func() {
		close(closeChan)
		fmt.Println("Closing CEC connection...")
	}()

	go func() {
		for {
			select {
			case <-closeChan:
				fmt.Println("Closing CEC connection...")
				return
			case e := <-cec.CallbackEvents:
				keyPressed, ok := e.(cec.KeyPress)
				if !ok {
					continue
				}
				fmt.Println("Got event:", keyPressed.KeyCode)
			}
		}
	}()

	c, err := cec.Open(adapter, device, "recording")
	panicOnError(err, "Failed to open CEC connection")

	defer c.Destroy()

	fmt.Println("CEC connection opened successfully.")
	time.Sleep(60 * time.Second)
	fmt.Println("You can now use CEC features.")
}

func checkUinput() {
	keyboard, err := uinput.CreateKeyboard()
	panicOnError(err, "Failed to create uinput keyboard")

	defer keyboard.Close()

	panicOnError(keyboard.KeyPress(uinput.KeyDown), "Failed to press key")
	time.Sleep(5 * time.Second)
}

func panicOnError(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", msg, err))
	}
}

func printHelp() {
	fmt.Println("Usage: go run main.go <command>")
	fmt.Println("Commands:")
	fmt.Println("  cec     - Check if CEC is available")
	fmt.Println("  uinput  - Check if uinput is available")
	fmt.Println("  hi      - Print 'HI'")
}
