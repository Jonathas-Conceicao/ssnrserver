package main

import (
	"log"
	"net"
	"os"

	"github.com/urfave/cli"

	ssnr "github.com/Jonathas-Conceicao/ssnrgo"
)

var users *ssnr.UserTable

func main() {
	app := cli.NewApp()
	app.Name = "SSNR server"
	app.Usage = "Host distributed notifications over SSNR protocol"
	app.Version = "0.1.0"

	cli.HelpFlag = cli.BoolFlag{
		Name:  "help",
		Usage: "show this dialog",
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "port, p",
			Value: ":30106",
			Usage: "Server port",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "Server's name",
		},
	}

	app.Action = func(c *cli.Context) error {
		log.Println("Allocating new Users Table")
		users = new(ssnr.UserTable)
		users.Add(0, ssnr.User{"Server", nil})

		config, err := ssnr.NewConfig(
			"Server",
			c.String("port"),
			c.String("name"))
		if err != nil {
			return err
		}
		tcpCon, err := startConnection(config)
		if err != nil {
			return err
		}

		for {
			conn, err := tcpCon.Accept()
			if err != nil {
				return err
			}
			go handleConnection(conn)
		}
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func startConnection(config *ssnr.Config) (net.Listener, error) {
	log.Println("Oppening TCP connection on port" + config.Port)
	r, err := net.Listen("tcp", config.Port)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func handleConnection(cn net.Conn) error {
	tmp := make([]byte, 500)
	_, err := cn.Read(tmp)
	if err != nil {
		return err
	}

	switch tmp[0] {
	case ssnr.NotificationCode:
		return handleNotification(tmp)
	case ssnr.ListingCode:
		return handleListing(cn, tmp)
	default:
		return handleUnknown(cn, tmp)
	}
}

func handleNotification(data []byte) error {
	message := ssnr.DecodeNotification(data)
	log.Println("Message Received:\n" + message.String())
	log.Println("Current list of users: ", users.Length())
	log.Print(users)
	return nil
}

func handleListing(cn net.Conn, data []byte) error {
	log.Print("Recived listing request")
	listing := ssnr.DecodeListing(data)
	log.Println(" for ", listing.GetAmount(), " users")
	listing.SetUsers(users)
	answer := listing.Encode()
	cn.Write(answer)
	return nil
}

func handleUnknown(cn net.Conn, data []byte) error {
	log.Printf("Invalid message received!\nCode: %d\n", data[0])
	cn.Write([]byte{0})
	return nil
}
