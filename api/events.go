package api

import "github.com/saurabh-newera/BLE/bluez/profile"

//EventStatus indicate the status related to an event
type EventStatus int

// DeviceStatus indicate the status of a device
type DeviceStatus EventStatus

const (
	//DeviceAdded indicates the device interface appeared
	DeviceAdded DeviceStatus = iota
	//DeviceRemoved indicates the device interface disappeared
	DeviceRemoved
)

const (
	// StatusAdded something has been added
	StatusAdded EventStatus = iota
	// StatusRemoved something has been removed
	StatusRemoved
)

//DiscoveredDeviceEvent contains detail regarding an ongoing discovery operation
type DiscoveredDeviceEvent struct {
	Path   string
	Status DeviceStatus
	Device *Device
}

// AdapterEvent reports the availability of a bluetooth adapter
type AdapterEvent struct {
	Name   string
	Path   string
	Status DeviceStatus
}

// PropertyChangedEvent an object to describe a changed property
type PropertyChangedEvent struct {
	Iface      string
	Field      string
	Value      interface{}
	Properties *profile.Device1Properties
	Device     *Device
}

// GattServiceEvent triggered when a new GattService1 is added or removed
type GattServiceEvent struct {
	Path       string
	DevicePath string
	Properties *profile.GattService1Properties
	Status     EventStatus
}

// GattCharacteristicEvent triggered when a new GattCharacteristic1 is added or removed
type GattCharacteristicEvent struct {
	Path       string
	DevicePath string
	Properties *profile.GattCharacteristic1Properties
	Status     EventStatus
}

// GattDescriptorEvent triggered when a new GattDescriptor1 is added or removed
type GattDescriptorEvent struct {
	Path       string
	DevicePath string
	Properties *profile.GattDescriptor1Properties
	Status     EventStatus
}

// DataEvent triggered when a new data value is available
type DataEvent struct {

	Device 						*Device
	SensorType					string
	
	AmbientTempValue 			interface{}
	AmbientTempUnit  			string
	
	ObjectTempValue   			interface{}
	ObjectTempUnit   			string
	
	SensorId	   				string
	
	BarometericPressureValue 	interface{}
	BarometericPressureUnit  	string
	
	BarometericTempValue  		interface{}
	BarometericTempUnit  		string
	
	HumidityValue				interface{}
	HumidityUnit				string

	HumidityTempValue			interface{}
	HumidityTempUnit			string

	MpuGyroscopeValue			interface{}
	MpuGyroscopeUnit			string
	
	MpuAccelerometerValue		interface{}
	MpuAccelerometerUnit		string
	
	MpuMagnetometerValue		interface{}
	MpuMagnetometerUnit			string
	
	LuxometerValue 				interface{}
	LuxometerUnit				string
	
	FirmwareVersion				string
	HardwareVersion				string
	Manufacturer				string
	Model						string
}