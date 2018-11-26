package main

import (
	"fmt"
	"go/build"
	"net"

	ssnr "github.com/Jonathas-Conceicao/ssnrgo"
)

var users *ssnr.UserTable

func main() {
	confFile := build.Default.GOPATH + "/configs/ssnr_server_config.json"
	config := loadConfig(confFile)
	loadUserTable()
	tcpCon := startConnection(config)

	for {
		conn, err := tcpCon.Accept()
		if err != nil {
			panic("Failed to accept TCP connection")
		}
		go handleConnection(conn)
	}
}

func loadConfig(filePath string) *ssnr.Config {
	fmt.Println("Loading config")
	r := ssnr.NewConfig(filePath)
	return r
}

func loadUserTable() {
	fmt.Println("Allocating new Users Table")
	users = new(ssnr.UserTable)
	users.Add(0, ssnr.User{"Server", nil})
}

func startConnection(config *ssnr.Config) net.Listener {
	fmt.Println("Oppening TCP connection on port" + config.Port)
	r, err := net.Listen("tcp", config.Port)
	if err != nil {
		panic("Failed to open TCP port at" + config.Port)
	}
	return r
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
