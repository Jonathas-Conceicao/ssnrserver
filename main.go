package main

import (
	"fmt"
	"net"

	ssnr "github.com/Jonathas-Conceicao/ssnrgo"
)

func main() {
	port := ":8080"
	fmt.Println("Oppening TCP connection on port" + port)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		panic("Failed to open TCP port at" + port)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	tmp := make([]byte, 500)
	_, err := conn.Read(tmp)
	if err != nil {
		panic("Error found when reading message")
	}

	switch tmp[0] {
	case 78:
		handleNotification(conn, tmp)
	default:
		handleUnknown(conn, tmp)
	}
	return
}

func handleNotification(conn net.Conn, data []byte) {
	message := ssnr.DecodeNotification(data)
	fmt.Println("Message Received:\n" + message.String())
	// sample process for string received
	conn.Write([]byte("Replied\n"))
	return
}

func handleUnknown(conn net.Conn, data []byte) {
	fmt.Printf("Invalid message received with code: %d", data[0])
	conn.Write([]byte("Invalid Message\n"))
	return
}
