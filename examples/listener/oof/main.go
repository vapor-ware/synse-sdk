package main

import (
	"net"
	"fmt"
)

func main() {

	//s, err := net.ResolveUDPAddr("udp", ":5884")
	//c, err := net.DialUDP("udp", nil, s)

	c, err := net.Dial("udp4", "localhost:5884")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
	defer c.Close()

	for {
		buffer := make([]byte, 1024)
		n, err := c.Read(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Reply: %s\n", string(buffer[0:n]))
	}
}
