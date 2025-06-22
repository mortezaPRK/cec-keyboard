package cec

import (
	"log"
	"log/slog"
)

// Open - open a new connection to the CEC device with the given name
func Open(name, deviceName, deviceType string) (*Connection, error) {
	c := new(Connection)

	var err error

	c.connection, err = cecInit(deviceName, deviceType)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	slog.Debug("CEC connection initialized",
		"device", deviceName,
		"type", deviceType,
		"name", name,
	)

	adapter, err := getAdapter(c.connection, name)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	slog.Debug("CEC adapter found")

	err = openAdapter(c.connection, adapter)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	slog.Debug("CEC adapter opened")

	return c, nil
}
