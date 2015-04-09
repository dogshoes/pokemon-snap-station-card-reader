package main

import (
        "log"
        "github.com/tarm/serial"
	"fmt"
	"bytes"
	"encoding/binary"
)

type Message struct {
	ack, ln, lrc byte
	data []byte
}

func read_message(s *serial.Port) (string, error) {
	buf := make([]byte, 1)
	var message [1000]byte
	count := 0

	for {
        	_, readerr := s.Read(buf)
        	if readerr != nil {
        	        return "", readerr
        	}

		if buf[0] == 0x03 {
			break
		}

		message[count] = buf[0]
		count = count + 1
	}

	return string(message[:count]), nil
}

func append_byte_to_message(buffer *bytes.Buffer, b byte) {
	hex := fmt.Sprintf("%X", b)

	// %X omits the first 0 for single byte characters.  Add it in.
	if len(hex) < 2 {
		buffer.WriteString("0")
	}

	buffer.WriteString(hex)
}

func construct_message(message ...byte) ([]byte) {
	var buffer bytes.Buffer
	var lrc byte
	size := len(message)

	// ACK
	append_byte_to_message(&buffer, 0x60)
	lrc = lrc ^ 0x60

	// We only want the least significant byte, messages can only be 255 characters long.
	lenbyte := make([]byte, 4)
	binary.BigEndian.PutUint32(lenbyte, uint32(size))

	// LN
	append_byte_to_message(&buffer, byte(lenbyte[3]))
	lrc = lrc ^ lenbyte[3]

	// <data>
	for i := 0; i < size; i++ {
		append_byte_to_message(&buffer, byte(message[i]))
		lrc = lrc ^ message[i]
	}

	// LRC
	append_byte_to_message(&buffer, byte(lrc))

	// EOT
	buffer.WriteString("\x03")

	return buffer.Bytes()
}

func main() {
        c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}
        s, err := serial.OpenPort(c)
        if err != nil {
                log.Fatal(err)
        }

	log.Println("Opened /dev/ttyUSB0.")

	log.Println(string(construct_message(0x17, 0x07)))

	_, err = s.Write(construct_message(0x17, 0x07))
        if err != nil {
                log.Fatal(err)
        }

	log.Println("Set sense type.")

	_, err = s.Write(construct_message(0x17))
        if err != nil {
                log.Fatal(err)
        }

	log.Println("Wrote inquiry.")


	message, readerr := read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(message)

	_, err = s.Write(construct_message(0x12))
        if err != nil {
                log.Fatal(err)
        }

	log.Println("Power up!")
	
	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(message)

	_, err = s.Write(construct_message(0x13, 0x00, 0xB0, 0x00, 0x02, 0x06))
        if err != nil {
                log.Fatal(err)
        }

	log.Println("Serial number inquiry.")

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(message)

	_, err = s.Write(construct_message(0x13, 0x00, 0xB2, 0x05, 0x08, 0x02))
        if err != nil {
                log.Fatal(err)
        }

	log.Println("Counter inquiry.")

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(message)

	_, err = s.Write(construct_message(0x14, 0x00, 0xD0, 0x00, 0x0C, 0x01, 0x01))
        if err != nil {
                log.Fatal(err)
        }

	log.Println("Add credit!")

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(message)

	_, err = s.Write(construct_message(0x13, 0x00, 0xB0, 0x00, 0x00, 0x10))
        if err != nil {
                log.Fatal(err)
        }

	log.Println("Memory inquiry.")

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(message)
}
