package main

import (
	"fmt"
	"github.com/saurabh-newera/BLE/linux"
)

var adapterID = "hci0"

func main() {

	hciconfig := linux.HCIConfig{}
	res, err := hciconfig.Up()
	if err != nil {
		panic(err)
	}
	fmt.Sprintf("Address %s, enabled %t", res.Address, res.Enabled)

	res, err = hciconfig.Down()
	if err != nil {
		panic(err)
	}
	fmt.Sprintf("Address %s, enabled %t", res.Address, res.Enabled)

}
