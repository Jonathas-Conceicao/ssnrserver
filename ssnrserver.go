package main

import (
	"fmt"
	"net"

	ssnr "github.com/Jonathas-Conceicao/ssnrgo"
)

var users *ssnr.UserTable

func main() {
	port := ":8080"
	fmt.Println("Oppening TCP connection on port" + port)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		panic("Failed to open TCP port at" + port)
	}

	fmt.Println("Allocating new Users Table")
	users = new(ssnr.UserTable)
	users.Add(78, ssnr.User{"Dummy01", nil})
	users.Add(79, ssnr.User{"Dummy02", nil})
	users.Add(80, ssnr.User{"Dummy03", nil})

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic("Failed to accept TCP connection")
		}
		go handleConnection(conn)
	}
}

func handleConnection(cn net.Conn) {
	tmp := make([]byte, 500)
	_, err := cn.Read(tmp)
	if err != nil {
		panic("Error found when reading message")
	}

	switch tmp[0] {
	case ssnr.NotificationCode:
		handleNotification(tmp)
	case ssnr.ListingCode:
		handleListing(cn, tmp)
	default:
		handleUnknown(cn, tmp)
	}
	return
}

func handleNotification(data []byte) {
	message := ssnr.DecodeNotification(data)
	fmt.Println("Message Received:\n" + message.String())
	fmt.Println("Current list of users: ", users.Length())
	fmt.Print(users)
	return
}

func handleListing(cn net.Conn, data []byte) {
	fmt.Print("Recived listing request")
	listing := ssnr.DecodeListing(data)
	fmt.Println(" for ", listing.GetAmount(), " users")
	listing.SetUsers(users)
	answer := listing.Encode()
	cn.Write(answer)
	return
}

func handleUnknown(cn net.Conn, data []byte) {
	fmt.Printf("Invalid message received!\nCode: %d\n", data[0])
	cn.Write([]byte{0})
	return
}
