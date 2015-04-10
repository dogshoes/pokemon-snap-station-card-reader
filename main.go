package main

import (
        "log"
        "github.com/tarm/serial"
	"fmt"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"os"
	"time"
	"errors"
)

const EOT = 0x03
const WriteDelay = 100 * time.Millisecond
const GPM103Type = 0x07

type Message struct {
	ack, ln, lrc byte
	data []byte
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal(fmt.Sprintf("Usage: %s [device]", os.Args[0]))
	}

        c := &serial.Config{Name: os.Args[1], Baud: 9600}
        s, err := serial.OpenPort(c)
        if err != nil {
                log.Fatal(err)
        }

	log.Println(fmt.Sprintf("Opened %s.", os.Args[1]))

	// Instruct the reader we'll be using a GPM103 card.
	log.Println(fmt.Sprintf("Setting sense type for GPM103 (0x%X).", GPM103Type))

	err = write_message(construct_message(0x17, GPM103Type), s)
        if err != nil {
                log.Fatal(err)
        }

	message, readerr := read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	// Inquire whether a card is present.
	log.Println("Sense card.")
	err = write_message(construct_message(0x17), s)
        if err != nil {
                log.Fatal(err)
        }

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	powered := false
	inserted := false

	// b0000001 sets whether the card is 3V.
	if (message.data[1] & 1) > 0 {
		log.Println(fmt.Sprintf("\tCard is 3V."))
	} else {
		log.Println(fmt.Sprintf("\tCard is 5V."))
	}

	// b00000010 sets whether the card is powered.
	if (message.data[1] & 2) > 0 {
		log.Println(fmt.Sprintf("\tCard is powered."))
		powered = true
	} else {
		log.Println(fmt.Sprintf("\tCard is not powered."))
	}

	// b00000100 sets whether the card is inserted.
	if (message.data[1] & 4) > 0 {
		log.Println(fmt.Sprintf("\tCard is inserted."))
		inserted = true
	} else {
		log.Println(fmt.Sprintf("\tCard is not inserted."))
	}

	// b00001000 sets whether we're T=1
	if (message.data[1] & 8) > 0 {
		log.Println(fmt.Sprintf("\tCard protocol is T=1."))
	} else {
		log.Println(fmt.Sprintf("\tCard protocol is T=0."))
	}

	if message.data[2] == GPM103Type {
		log.Println(fmt.Sprintf("\tDetected card of type 0x%X.", message.data[2]))
	} else {
		log.Println(fmt.Sprintf("\tDetected card of type %X.  Expected 0x%X!!!!", message.data[2], GPM103Type))
	}

	if inserted == false {
		log.Println("No card detected.")
		os.Exit(0)
	}

	// Power the card up.
	if powered == false {
		log.Println("Powering up the card.")

		err = write_message(construct_message(0x12), s)
		if err != nil {
		        log.Fatal(err)
		}
	
		message, readerr = read_message(s)
		if readerr != nil {
		        log.Fatal(readerr)
		}

		log.Println(fmt.Sprintf("\tCard ATR'd with 0x%X", message.data[1:]))
	}

	// Read the card's serial number.
	log.Println("Serial number inquiry.")

	err = write_message(construct_message(0x13, 0x00, 0xB0, 0x00, 0x02, 0x06), s)
        if err != nil {
                log.Fatal(err)
        }

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(fmt.Sprintf("\t0x%X", message.data[1:7]))

	// Read the card's counter.
	log.Println("Counter inquiry.")

	err = write_message(construct_message(0x13, 0x00, 0xB2, 0x05, 0x08, 0x02), s)
        if err != nil {
                log.Fatal(err)
        }

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(fmt.Sprintf("\tCounter is 0x%X credits.", message.data[1:3]))

	/* Writing to the counter doesn't work, we receive an error about invalid instruction code.
	log.Println("Writing to counter.")

	err = write_message(construct_message(0x14, 0x00, 0xD2, 0x05, 0x08, 0x02, 0x00, 0x02), s)
        if err != nil {
                log.Fatal(err)
        }

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(fmt.Sprintf("\t%X", message.data))
	*/

	/* Writing to memory doesn't return an error but also doesn't do anything.
	log.Println("Writing to memory (at counter position).")

	err = write_message(construct_message(0x14, 0x00, 0xD0, 0x00, 0x0C, 0x01, 0x01), s)
        if err != nil {
                log.Fatal(err)
        }

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(fmt.Sprintf("%X", message.data))
	*/

	/*
	log.Println("Dumping memory...")

	// Memory wraps around after 0x10 bytes.
	err = write_message(construct_message(0x13, 0x00, 0xB0, 0x00, 0x00, 0x10), s)
        if err != nil {
                log.Fatal(err)
        }

	message, readerr = read_message(s)
	if readerr != nil {
                log.Fatal(readerr)
        }

	log.Println(fmt.Sprintf("\t0x%X", message.data))
	*/
}

func write_message(message []byte, s *serial.Port) (error) {
	_, err := s.Write(message)
        if err != nil {
                return err
        }

	time.Sleep(WriteDelay)
	
	return nil
}

func read_message(s *serial.Port) (*Message, error) {
	buf := make([]byte, 1)
	var rawmessage [1000]byte
	count := 0

	for {
        	_, readerr := s.Read(buf)
        	if readerr != nil {
        	        return nil, readerr
        	}

		// Have we reached the end of the message?
		if buf[0] == EOT {
			break
		}

		rawmessage[count] = buf[0]
		count = count + 1
	}

	if count % 2 != 0 {
		return nil, errors.New("Message length is not divisible by two.  Invalid message.")
	} else if count < 6 {
		return nil, errors.New("Received message smaller than allowable (3 bytes). Cannot decode.")
	}

	decodedlen := hex.DecodedLen(count)
	buffer := make([]byte, decodedlen)
	decodedbytes, err := hex.Decode(buffer, rawmessage[:count])

	if err != nil {
		return nil, err
	}

	message := new(Message)
	message.ack = buffer[0]
	message.ln = buffer[1]
	message.data = buffer[2:decodedbytes - 1]
	message.lrc = buffer[decodedbytes - 1]

	// Check the returned LRC to ensure it's valid and we have a complete message.
	var expectedlrc byte
	for i := 0; i < len(buffer) - 2; i++ {
		expectedlrc = expectedlrc ^ buffer[i]
	}

	if expectedlrc != message.lrc {
		return nil, errors.New("Received message fails checksum.")
	}

	return message, nil
}

func append_byte_to_message(buffer *bytes.Buffer, b byte) {
	hex := fmt.Sprintf("%X", b)

	// %X omits the first 0 for single byte characters.  Add it in.
	if len(hex) < 2 {
		buffer.WriteString("0")
	}

	buffer.WriteString(hex)
}

func int_to_byte(i uint32) (byte) {
	lenbyte := make([]byte, 4)
	binary.BigEndian.PutUint32(lenbyte, i)
	return byte(lenbyte[3])
}

func construct_message(message ...byte) ([]byte) {
	var buffer bytes.Buffer
	var lrc byte
	size := len(message)

	// ACK
	append_byte_to_message(&buffer, 0x60)
	lrc = lrc ^ 0x60

	// LN
	ln := int_to_byte(uint32(size))
	append_byte_to_message(&buffer, ln)
	lrc = lrc ^ ln

	// <data>
	for i := 0; i < size; i++ {
		append_byte_to_message(&buffer, byte(message[i]))
		lrc = lrc ^ message[i]
	}

	// LRC
	append_byte_to_message(&buffer, byte(lrc))

	// EOT
	buffer.WriteByte(EOT)

	return buffer.Bytes()
}
