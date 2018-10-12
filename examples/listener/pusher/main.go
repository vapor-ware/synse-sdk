package main

import (
	"encoding/binary"
	"log"
	"math/rand"
	"net"
	"time"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":8553")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.Dial("udp", addr.String())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // nolint: errcheck

	log.Printf("Sending data on: %v", addr.String())
	for {
		b := make([]byte, 4)
		data := rand.Uint32()
		binary.LittleEndian.PutUint32(b, data)
		log.Printf("<< %v", data)
		_, err := conn.Write(b)
		if err != nil {
			log.Printf("failed to write. continuing.")
		}
		time.Sleep(3 * time.Second)
	}
}
