package main

import (
	"fmt"
	"github.com/godbus/dbus"
)

// read sensor tag temperature
func testrun() {

	conn, err := dbus.SystemBus()
	if err != nil {
		panic(err)
	}

	ns := "org.bluez"
	//
	temperatureDataPath := "/org/bluez/hci0/dev_B0_B4_48_C9_4B_01/service0022/char0023"
	temperatureConfigPath := "/org/bluez/hci0/dev_B0_B4_48_C9_4B_01/service0022/char0026"

	writeCall := "org.bluez.GattCharacteristic1.WriteValue"
	readCall := "org.bluez.GattCharacteristic1.ReadValue"

	// write, enable measurements

	opts := make(map[string]dbus.Variant)

	fmt.Sprintf("Enable measurment")

	b := []byte{0x01}

	temperatureConfig := conn.Object(ns, dbus.ObjectPath(temperatureConfigPath))
	call := temperatureConfig.Call(writeCall, 0, b, opts)
	if call.Err != nil {
		fmt.Sprintf("Error: %s", call.Err)
		panic(call.Err)
	}

	// read
	fmt.Sprintf("Read data")
	temperatureData := conn.Object(ns, dbus.ObjectPath(temperatureDataPath))
	call = temperatureData.Call(readCall, 0, opts)
	if call.Err != nil {
		fmt.Sprintf("Error: %s", call.Err)
		panic(call.Err)
	}

	fmt.Sprintf("Result %v", call.Body)

	fmt.Sprintf("Disable measurment")

	b = []byte{0x00}
	call = temperatureConfig.Call(writeCall, 0, b, opts)
	if call.Err != nil {
		fmt.Sprintf("Error: %s", call.Err)
		panic(call.Err)
	}

}
