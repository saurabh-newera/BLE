package api

import (
	"errors"
	"reflect"
	"strings"

	"github.com/godbus/dbus"
	"github.com/saurabh-newera/BLE/bluez"
	"github.com/saurabh-newera/BLE/bluez/profile"
	"github.com/saurabh-newera/BLE/emitter"
	"github.com/saurabh-newera/BLE/util"
	"github.com/tj/go-debug"
	logging "github.com/op/go-logging"
	"fmt"
)
var log = logging.MustGetLogger("examples")
var dbgDevice = debug.Debug("bluez:api:Device")
var deviceRegistry = make(map[string]*Device)

// NewDevice creates a new Device
func NewDevice(path string) *Device {

	if _, ok := deviceRegistry[path]; ok {
		// dbgDevice("Reusing cache instance %s", path)
		return deviceRegistry[path]
	}

	d := new(Device)
	d.Path = path
	d.client = profile.NewDevice1(path)

	d.client.GetProperties()
	d.Properties = d.client.Properties
	d.chars = make(map[dbus.ObjectPath]*profile.GattCharacteristic1, 0)

	deviceRegistry[path] = d

	// d.watchProperties()

	return d
}

//ClearDevice free a device struct
func ClearDevice(d *Device) {

	d.Disconnect()
	d.unwatchProperties()
	c, err := d.GetClient()
	if err == nil {
		c.Close()
	}

	if _, ok := deviceRegistry[d.Path]; ok {
		delete(deviceRegistry, d.Path)
	}

}

// ParseDevice parse a Device from a ObjectManager map
func ParseDevice(path dbus.ObjectPath, propsMap map[string]dbus.Variant) (*Device, error) {

	d := new(Device)
	d.Path = string(path)
	d.client = profile.NewDevice1(d.Path)

	props := new(profile.Device1Properties)
	util.MapToStruct(props, propsMap)
	c, err := d.GetClient()
	if err != nil {
		return nil, err
	}
	c.Properties = props

	return d, nil
}

func (d *Device) watchProperties() error {

	dbgDevice("watch-prop: watching properties")

	channel, err := d.client.Register()
	if err != nil {
		return err
	}

	go (func() {
		for {

			if channel == nil {
				dbgDevice("watch-prop: nil channel, exit")
				break
			}

			dbgDevice("watch-prop: waiting updates")
			sig := <-channel

			if sig == nil {
				dbgDevice("watch-prop: nil sig, exit")
				return
			}

			dbgDevice("watch-prop: signal name %s", sig.Name)
			if sig.Name != bluez.PropertiesChanged {
				dbgDevice("Skipped %s vs %s\n", sig.Name, bluez.PropertiesInterface)
				continue
			}

			dbgDevice("Device property changed")
			for i := 0; i < len(sig.Body); i++ {
				dbgDevice("%s -> %s", reflect.TypeOf(sig.Body[i]), sig.Body[i])
			}

			iface := sig.Body[0].(string)
			changes := sig.Body[1].(map[string]dbus.Variant)
			for field, val := range changes {

				// updates [*]Properties struct
				props := d.Properties

				s := reflect.ValueOf(props).Elem()
				// exported field
				f := s.FieldByName(field)
				if f.IsValid() {
					// A Value can be changed only if it is
					// addressable and was not obtained by
					// the use of unexported struct fields.
					if f.CanSet() {
						x := reflect.ValueOf(val.Value())
						f.Set(x)
						dbgDevice("Set props value: %s = %s\n", field, x.Interface())
					}
				}

				dbgDevice("Emit change for %s = %v\n", field, val.Value())
				propChanged := PropertyChangedEvent{string(iface), field, val.Value(), props, d}
				d.Emit("changed", propChanged)
			}
		}
	})()

	return nil
}

//Device return an API to interact with a DBus device
type Device struct {
	Path       string
	Properties *profile.Device1Properties
	client     *profile.Device1
	chars      map[dbus.ObjectPath]*profile.GattCharacteristic1
}

func (d *Device) unwatchProperties() error {
	return d.client.Unregister()
}

//GetClient return a DBus Device1 interface client
func (d *Device) GetClient() (*profile.Device1, error) {
	if d.client == nil {
		return nil, errors.New("Client not available")
	}
	return d.client, nil
}

//GetProperties return the properties for the device
func (d *Device) GetProperties() (*profile.Device1Properties, error) {

	if d == nil {
		return nil, errors.New("Empty device pointer")
	}

	c, err := d.GetClient()
	if err != nil {
		return nil, err
	}

	props, err := c.GetProperties()

	if err != nil {
		return nil, err
	}

	d.Properties = props
	return d.Properties, err
}

//GetProperty return a property value
func (d *Device) GetProperty(name string) (data interface{}, err error) {
	c, err := d.GetClient()
	if err != nil {
		return nil, err
	}
	val, err := c.GetProperty(name)
	if err != nil {
		return nil, err
	}
	return val.Value(), nil
}

//On register callback for event
func (d *Device) On(name string, fn *emitter.Callback) {
	switch name {
	case "changed":
		d.watchProperties()
		break
	}
	emitter.On(d.Path+"."+name, fn)
}

//Off unregister callback for event
func (d *Device) Off(name string, cb *emitter.Callback) {
	switch name {
	case "changed":
		d.unwatchProperties()
		break
	}

	pattern := d.Path + "." + name
	if name != "*" {
		emitter.Off(pattern, cb)
	} else {
		emitter.RemoveListeners(pattern, nil)
	}
}

//Emit an event
func (d *Device) Emit(name string, data interface{}) {
	emitter.Emit(d.Path+"."+name, data)
}

//GetService return a GattService
func (d *Device) GetService(path string) *profile.GattService1 {
	return profile.NewGattService1(path)
}

//GetChar return a GattService
func (d *Device) GetChar(path string) *profile.GattCharacteristic1 {
	return profile.NewGattCharacteristic1(path)
}

//..........................................................................................

//GetAllServicesAndUUID return a list of uuid's with their corresponding services.....

func (d *Device) GetAllServicesAndUUID() ([]string, error) {

	list := d.GetCharsList()

	//log.Debug("Find by uuid, char list %d, cached list %d", len(list), len(d.chars))
	
	var deviceFound []string
	var uuidAndService string
	for _, path := range list {
		
		// use cache
		_, ok := d.chars[path]
		if !ok {
			d.chars[path] = profile.NewGattCharacteristic1(string(path))
		}
		
		props := d.chars[path].Properties
		cuuid := strings.ToUpper(props.UUID)
		service := string(props.Service)

		uuidAndService = fmt.Sprint(cuuid,":",service)
		deviceFound = append(deviceFound, uuidAndService)
	}

	if deviceFound == nil {
		//dbgDevice("Characteristic not Found: %s ", uuid)
	}

	return deviceFound,nil
}
//.....................................................................

//GetCharByUUID return a GattService by its uuid, return nil if not found
func (d *Device) GetCharByUUID(uuid string) (*profile.GattCharacteristic1, error) {

	uuid = strings.ToUpper(uuid)

	list := d.GetCharsList()

	dbgDevice("Find by uuid, char list %d, cached list %d", len(list), len(d.chars))

	var deviceFound *profile.GattCharacteristic1

	for _, path := range list {

		// use cache
		_, ok := d.chars[path]
		if !ok {
			d.chars[path] = profile.NewGattCharacteristic1(string(path))
		}

		props := d.chars[path].Properties
		cuuid := strings.ToUpper(props.UUID)
		//log.Debug("Properties : ",props)
		if cuuid == uuid {
			dbgDevice("Found char %s", uuid)
			deviceFound = d.chars[path]
		}
	}

	if deviceFound == nil {
		dbgDevice("Characteristic not Found: %s ", uuid)
	}

	return deviceFound, nil
}

//GetCharsList return a device characteristics
func (d *Device) GetCharsList() []dbus.ObjectPath {

	var chars []dbus.ObjectPath

	if len(d.chars) != 0 {

		for objpath := range d.chars {
			chars = append(chars, objpath)
		}

		dbgDevice("Cached %d chars", len(chars))
		return chars
	}

	dbgDevice("Scanning chars by ObjectPath")
	list := GetManager().GetObjects()
	for objpath := range *list {
		path := string(objpath)
		if !strings.HasPrefix(path, d.Path) {
			continue
		}
		charPos := strings.Index(path, "char")
		if charPos == -1 {
			continue
		}
		if strings.Index(path[charPos:], "desc") != -1 {
			continue
		}

		chars = append(chars, objpath)
	}

	dbgDevice("Found %d chars", len(chars))
	return chars
}

//IsConnected check if connected to the device
func (d *Device) IsConnected() bool {

	props, _ := d.GetProperties()

	if props == nil {
		return false
	}

	return props.Connected
}

//Connect to device
func (d *Device) Connect() error {

	c, err := d.GetClient()
	if err != nil {
		return err
	}

	err = c.Connect()
	if err != nil {
		return err
	}
	return nil
}

//Disconnect from a device
func (d *Device) Disconnect() error {
	c, err := d.GetClient()
	if err != nil {
		return err
	}
	c.Disconnect()
	return nil
}

//Pair a device
func (d *Device) Pair() error {
	c, err := d.GetClient()
	if err != nil {
		return err
	}
	c.Pair()
	return nil
}
