package cec

/*
#cgo pkg-config: libcec
#include <stdlib.h>
#include <stdio.h>
#include <errno.h>
#include <libcec/cecc.h>

ICECCallbacks g_callbacks;
// callbacks.go exports
void keyPressCallback(void *, const cec_keypress*);

void setupCallbacks(libcec_configuration *conf)
{
	g_callbacks.logMessage = NULL;
	g_callbacks.keyPress = &keyPressCallback;
	g_callbacks.commandReceived = NULL;
	g_callbacks.configurationChanged = NULL;
	g_callbacks.alert = NULL;
	g_callbacks.menuStateChanged = NULL;
	g_callbacks.sourceActivated = NULL;
	(*conf).callbacks = &g_callbacks;
}

void setName(libcec_configuration *conf, char *name)
{
	snprintf((*conf).strDeviceName, 13, "%s", name);
}

*/
import "C"

import (
	"errors"
	"strings"
	"unsafe"
)

var deviceTypes = map[string]C.cec_device_type{
	"tv":        C.CEC_DEVICE_TYPE_TV,
	"recording": C.CEC_DEVICE_TYPE_RECORDING_DEVICE,
	"reserved":  C.CEC_DEVICE_TYPE_RESERVED,
	"tuner":     C.CEC_DEVICE_TYPE_TUNER,
	"playback":  C.CEC_DEVICE_TYPE_PLAYBACK_DEVICE,
	"audio":     C.CEC_DEVICE_TYPE_AUDIO_SYSTEM,
}

// Connection class
type Connection struct {
	connection C.libcec_connection_t
}

type cecAdapter struct {
	Path string
	Comm string
}

func cecInit(deviceName, deviceType string) (C.libcec_connection_t, error) {
	var connection C.libcec_connection_t
	var conf C.libcec_configuration

	conf.clientVersion = C.uint32_t(C.LIBCEC_VERSION_CURRENT)

	for i := 0; i < 5; i++ {
		conf.deviceTypes.types[i] = C.CEC_DEVICE_TYPE_RESERVED
	}

	cecDeviceType, ok := deviceTypes[deviceType]
	if !ok {
		return C.libcec_connection_t(nil), errors.New("invalid device type: " + deviceType)
	}

	conf.deviceTypes.types[0] = cecDeviceType

	cDeviceName := C.CString(deviceName)
	defer C.free(unsafe.Pointer(cDeviceName))

	C.setName(&conf, cDeviceName)
	C.setupCallbacks(&conf)

	connection = C.libcec_initialise(&conf)
	if connection == C.libcec_connection_t(nil) {
		return connection, errors.New("failed to init CEC")
	}
	return connection, nil
}

func getAdapter(connection C.libcec_connection_t, name string) (cecAdapter, error) {
	var adapter cecAdapter

	var deviceList [10]C.cec_adapter
	devicesFound := int(C.libcec_find_adapters(connection, &deviceList[0], 10, nil))

	for i := 0; i < devicesFound; i++ {
		device := deviceList[i]
		adapter.Path = C.GoStringN(&device.path[0], 1024)
		adapter.Comm = C.GoStringN(&device.comm[0], 1024)

		if strings.Contains(adapter.Path, name) || strings.Contains(adapter.Comm, name) {
			return adapter, nil
		}
	}

	return adapter, errors.New("no Device Found")
}

func openAdapter(connection C.libcec_connection_t, adapter cecAdapter) error {
	C.libcec_init_video_standalone(connection)

	result := C.libcec_open(connection, C.CString(adapter.Comm), C.CEC_DEFAULT_CONNECT_TIMEOUT)
	if result < 1 {
		return errors.New("failed to open adapter")
	}

	return nil
}

// Destroy - destroy the cec connection
func (c *Connection) Destroy() {
	C.libcec_destroy(c.connection)
}
