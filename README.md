# CEC Keyboard Handler

This tool maps CEC (Consumer Electronics Control) key events to uinput keyboard events, allowing you to control your Linux system with a TV remote that supports the CEC protocol. This is mainly used for controlling a Raspberry Pi or similar devices connected to a TV via HDMI, but it should work with any Linux system that supports CEC and uinput.

## Usage

```
cec-keyboard [flags] -mapping <mapping>...
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

- `-adapter` (optional): Specify the CEC adapter name to use. If omitted, the default adapter will be used.
- `-name` (optional): Specify the CEC device name to use.
- `-type` (optional): Specify the CEC device type. Default is `recording`. Other options: tv,recording,tuner,playback,audio.
- `-mapping` (required, can be specified multiple times): Define key mappings directly on the command line.
- `-log-level` (optional): Set the logging level. Options: debug, info, warn, error. Default: info.

### Example

```
# Use default adapter and device with a single mapping
go run . -mapping 0=p:1 -mapping 103=p:103

# Use a specific adapter with multiple mappings and debug logging
go run . -adapter /dev/cec0 -mapping 0=p:1 -mapping 103=p:103 -log-level debug
```

## Mapping Format

The mapping format is as follows:
```
<CEC key code>=<action>:<key code>[,<action>:<key code>,...]
```

Where:
- `<CEC key code>` is the integer code from the CEC remote event
- `<action>` can be:
  - `p`: Press (key down and then up)
  - `d`: Key down only
  - `u`: Key up only
  - `h`: Hold (key down and key up stacked for later release)
- `<key code>` is the uinput key code to send

You can chain multiple actions for a single CEC key by separating them with commas.

### Examples

```
# Press Escape key when CEC code 0 is received
0=p:1

# Press Up arrow when CEC code 103 is received
103=p:103

# Press Down arrow followed by Enter when CEC code 108 is received
108=p:108,p:28

# Hold Alt and press Tab when CEC code 10 is received
10=d:56,p:15,u:56
```

### Where to Find CEC Key Codes
You can run the cli without any mappings and check the logs to see the CEC key codes that are received from your remote. The CEC key codes are typically integers that correspond to specific buttons on your remote control.

### Where to Find Uinput Key Codes
You can find the list of uinput key codes in the Linux kernel documentation or by checking the [uinput_defs.go](https://github.com/sashko/go-uinput/blob/c753d6644126b88b83f62844f29ee998e2bd3139/uinput_defs.go#L72) file in the go-uinput library.


