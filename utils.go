package accessory

import (
	"encoding/binary"
	"github.com/google/gousb"
	"math/rand"
)

// SkipList is a list of vendor IDs that are known to not support the accessory protocol
// and should be skipped
// TODO: Add more vendor IDs to the list
var SkipList = []uint16{
	0x8087, // Intel Corp.
	0x1d6b, // Linux Foundation
	0x2109, // VIA Labs, Inc.
}

// GetDevices return a list of devices that support the specified protocol version
func GetDevices(protocolVersion uint16) (accessoryList []AccessoryDevice, err error) {
	accessoryList = make([]AccessoryDevice, 0)
	devices, err := gousb.NewContext().OpenDevices(func(desc *gousb.DeviceDesc) bool {
		for _, id := range SkipList {
			if desc.Vendor == gousb.ID(id) {
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	waitChan := make(chan *AccessoryDevice, len(devices))
	for _, d := range devices {
		d := d
		go func() {
			p, err := getProtocol(d)
			if err != nil || p < protocolVersion {
				waitChan <- nil
				return
			}
			waitChan <- NewAccessoryDevice(d, p)
		}()
	}
	for _, d := range devices {
		d := d
		accessory := <-waitChan
		if accessory != nil {
			accessoryList = append(accessoryList, *accessory)
			continue
		}
		_ = d.Close()
	}
	return accessoryList, nil
}

// getProtocol return the protocol version of the device
func getProtocol(dev *gousb.Device) (protocol uint16, err error) {
	if dev == nil {
		return 0, ErrorNoAccessoryDevice
	}
	data := make([]byte, 2)
	_, err = dev.Control(RTypeIn, AccessoryGetProtocol, 0, 0, data)
	if err != nil {
		return 0, ErrorFailedToGetProtocol
	}
	protocol = binary.LittleEndian.Uint16(data)
	return
}

// uint16InList return the index of the value in the list
func uint16InList(list []uint16, value uint16) (int, bool) {
	for i, v := range list {
		if v == value {
			return i, true
		}
	}
	return -1, false
}

// uint16GetUniqueRandom return a random uint16 that is not in the list
func uint16GetUniqueRandom(list []uint16) uint16 {
	for {
		r := uint16GetRandom()
		if r <= 100 {
			continue
		}
		if _, ok := uint16InList(list, r); !ok {
			return r
		}
	}
}

// uint16GetRandom return a random uint16
func uint16GetRandom() uint16 {
	return uint16(rand.Intn(0xffff))
}
