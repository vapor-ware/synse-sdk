package main

import (
	"encoding/binary"
	"log"
	"math/rand"
	"net"
	"time"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp4", "localhost:8553")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // nolint: errcheck

	rand.Seed(time.Now().Unix())

	log.Printf("Sending data on: %v", addr.String())
	for {
		data := make([]byte, 4)
		val := rand.Uint32()
		binary.LittleEndian.PutUint32(data, val)
		_,  err := conn.WriteToUDP(data, addr)
		if err != nil {
			log.Print("failed to write, continuing")

		} else {
			log.Printf("<< %v\t%v", val, data)
		}
		time.Sleep(2 * time.Second)
	}
}
