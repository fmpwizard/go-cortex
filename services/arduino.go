package services

//Most code on this page is from http://reprage.com/post/using-golang-to-connect-raspberrypi-and-arduino/

import (
	"bytes"
	"encoding/binary"
	"github.com/huin/goserial"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func Arduino(command string, light int) {
	// Find the device that represents the arduino serial
	// connection.
	c := &goserial.Config{Name: findArduino(), Baud: 9600}
	log.Printf("the port we will use is %v\n\n", findArduino())
	s, _ := goserial.OpenPort(c)
	// When connecting to an older revision Arduino, you need to wait
	// a little while it resets.
	time.Sleep(2 * time.Second)
	var arduinoCmd byte
	if command == "on" {
		arduinoCmd = 'u'
	} else {
		arduinoCmd = 'd'
	}

	sendArduinoCommand(arduinoCmd, uint32(light), s) //u for up, d for down
}

// findArduino looks for the file that represents the Arduino
// serial connection. Returns the fully qualified path to the
// device if we are able to find a likely candidate for an
// Arduino, otherwise an empty string if unable to find
// something that 'looks' like an Arduino device.
func findArduino() string {
	contents, _ := ioutil.ReadDir("/dev")

	// Look for what is mostly likely the Arduino device
	for _, f := range contents {
		if strings.Contains(f.Name(), "tty.usbserial") ||
			strings.Contains(f.Name(), "ttyUSB") ||
			strings.Contains(f.Name(), "ttyACM") {
			return "/dev/" + f.Name()
		}
	}

	// Have not been able to find a USB device that 'looks'
	// like an Arduino.
	return ""
}

// sendArduinoCommand transmits a new command over the nominated serial
// port to the arduino. Returns an error on failure. Each command is
// identified by a single byte and may take one argument (a float).
func sendArduinoCommand(command byte, argument uint32, serialPort io.ReadWriteCloser) error {
	if serialPort == nil {
		return nil
	}

	// Package argument for transmission
	bufOut := new(bytes.Buffer)
	err := binary.Write(bufOut, binary.LittleEndian, argument)
	if err != nil {
		return err
	}

	// Transmit command and argument down the pipe.
	for _, v := range [][]byte{[]byte{command}, bufOut.Bytes()} {
		_, err = serialPort.Write(v)
		if err != nil {
			return err
		}
	}

	return nil
}
