package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/srizzling/gotham/libs"
)

// MACAddress type, registery currently supports only strings (maybe using protobuff can fix this)
type MACAddress [6]byte

// MagicPacket is a packet used for sending WOL to devices
type MagicPacket struct {
	Header  [6]byte
	Payload [16]MACAddress
}

func getDevice(alias string) (*libs.Device, error) {
	url := "http://localhost:8080/view/" + alias
	response, err := http.Get(url)
	var device libs.Device

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	jsonErr := json.Unmarshal(body, &device)

	if jsonErr != nil {
		return nil, err
	}

	return &device, nil
}

// GetMacAddress from an alias and return a pointer to a MacAddress
func getMacAddress(alias string) (*MACAddress, error) {
	//var ret *MACAddres
	// Convert the MacAddress string to a byte array from devices map
	var deviceMAC MACAddress

	// Assign a byte array to each object and store in map
	device, err := getDevice(alias)

	fmt.Println(device.Alias)

	hwAddr, err := net.ParseMAC(device.HWAddress)
	if err != nil {

		return nil, err
	}

	for idx := range deviceMAC {
		deviceMAC[idx] = hwAddr[idx]
	}

	return &deviceMAC, nil
}

// NewMagicPacket will craft a magic packet provided you a provide a valid macaddress
func newMagicPacket(mac *MACAddress) *MagicPacket {
	var packet MagicPacket

	// fill the header of the packet with 0xFF
	for i := range packet.Header {
		packet.Header[i] = 0xFF
	}

	// repeat the macaddress 16times in the payload
	for i := range packet.Payload {
		packet.Payload[i] = *mac
	}

	return &packet
}

// SendMagicPacket sends a packet to a given an alias
func sendMagicPacket(alias string) error {
	// Get a macAddr from string
	macAddr, err := getMacAddress(alias)

	if err != nil {
		return err
	}

	magicPacket := newMagicPacket(macAddr)
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, magicPacket)

	// Get a UDPAddr to send the broadcast to
	udpAddr, err := net.ResolveUDPAddr("udp", "")

	var localAddr *net.UDPAddr

	// Open a UDP connection, and defer it's cleanup
	connection, err := net.DialUDP("udp", localAddr, udpAddr)
	if err != nil {
		return errors.New("inability to dial UDP address: " + err.Error())
	}
	defer connection.Close()

	// Write the bytes of the MagicPacket to the connection
	bytesWritten, err := connection.Write(buf.Bytes())
	if err != nil {
		fmt.Printf("Unable to write packet to connection\n")
		return err
	} else if bytesWritten != 102 {
		fmt.Printf("Warning: %d bytes written, %d expected!\n", bytesWritten, 102)
	}

	return nil
}

func main() {
	sendMagicPacket("Batman2")
}
