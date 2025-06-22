# CEC Keyboard & Mouse Handler

This tool maps CEC (Consumer Electronics Control) key events to uinput keyboard and mouse events, allowing you to control your Linux system with a TV remote that supports the CEC protocol. This is mainly used for controlling a Raspberry Pi or similar devices connected to a TV via HDMI, but it should work with any Linux system that supports CEC and uinput.

## Usage

```
cec-keyboard [flags]
```

## Permissions

You need to have access to `/dev/uinput`. This can be achieved by:
1. Adding your user to the `input` group:
   ```bash
   sudo usermod -aG input $USER
   ```
   After running this command, you may need to log out and back in for the group change to take effect.

2. Ensuring the `/dev/uinput` device exists and is writable by your user. using udev rules:
   ```bash
   sudo tee /etc/udev/rules.d/99-uinput.rules <<EOF
   KERNEL=="uinput", GROUP="input", MODE="0660"
   EOF
   ```
   After creating the udev rule, reload the udev rules and trigger them:
   ```bash
   sudo udevadm control --reload-rules
   sudo udevadm trigger
   ```

### Flags

- `-config` (optional): Path to the configuration file. Default is `config.yaml`.
- `-log-level` (optional): Set the logging level. Options: debug, info, warn, error. Default: info.

### Example

```
# Run with default config.yaml file
cec-keyboard

# Use a specific config file with debug logging
cec-keyboard -config my-config.yaml -log-level debug
```

## Configuration File Format

The configuration file uses YAML format with the following structure:

```yaml
adapter: /dev/cec0  # Optional: CEC adapter name/path
name: cec-keyboard  # Optional: CEC device name
type: recording     # Optional: CEC device type (tv, recording, tuner, playback, audio)
mappings:
  - cecCode: 0      # CEC key code from remote
    actions:        # List of actions to perform, can be multiple
      - keyboard:   # Keyboard action
          type: press  # Action type (press, down, up)
          code: 1      # uinput key code (1 is Escape)
      
  - cecCode: 103
    actions:
      - keyboard:
          type: press
          code: 103   # Up arrow key

  - cecCode: 108    # Can have multiple actions for one CEC code
    actions:
      - keyboard:
          type: press
          code: 108   # Down arrow key
      - keyboard:
          type: press
          code: 28    # Enter key
          
  # Mouse actions example
  - cecCode: 10
    actions:
      - mouse:
          leftClick: {} # Perform left mouse click
          
  - cecCode: 11
    actions:
      - mouse:
          rightClick: {}  # Perform right mouse click
          
  - cecCode: 12
    actions:
      - mouse:
          moveX: 10  # Move mouse 10 pixels right
          
  - cecCode: 13
    actions:
      - mouse:
          moveY: -10 # Move mouse 10 pixels up
```

### Action Types

#### Keyboard Actions
- `press`: Simulates pressing and releasing a key
- `down`: Simulates pressing a key down (without releasing)
- `up`: Simulates releasing a key

#### Mouse Actions
- `leftClick`: Simulates a left mouse button click
- `rightClick`: Simulates a right mouse button click
- `middleClick`: Simulates a middle mouse button click
- `sideClick`: Simulates a side mouse button click
- `moveX`: Moves the mouse cursor horizontally (positive value for right, negative for left)
- `moveY`: Moves the mouse cursor vertically (positive value for down, negative for up)

You can define multiple actions for the same CEC key code, allowing for complex control schemes.

### Where to Find CEC Key Codes
You can run the application with an empty or minimal mapping configuration and check the logs to see the CEC key codes that are received from your remote. Set the log level to debug to see additional information:

```
cec-keyboard -log-level debug
```

The CEC key codes are typically integers that correspond to specific buttons on your remote control.

### Where to Find Uinput Key Codes
You can find the list of uinput key codes in the Linux kernel documentation or by checking the [uinput_defs.go](https://github.com/sashko/go-uinput/blob/c753d6644126b88b83f62844f29ee998e2bd3139/uinput_defs.go#L72) file in the go-uinput library.

