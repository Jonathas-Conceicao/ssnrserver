package main

import (
	"bufio"
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
			log.Println("Connection Accepted from:", conn.RemoteAddr())
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
	reader := bufio.NewReader(cn)
	code, err := reader.Peek(1)
	if err != nil {
		return err
	}

	switch code[0] {
	case ssnr.NotificationCode:
		return handleNotification(cn, reader)
	case ssnr.ListingCode:
		return handleListing(cn, reader)
	case ssnr.RegisterCode:
		return handleRegister(cn, reader)
	case ssnr.DisconnectCode:
		return handleDisconnect(cn, reader)
	default:
		return handleUnknown(cn, reader)
	}
}

func handleNotification(cn net.Conn, rd *bufio.Reader) error {
	defer logAndClose(cn)
	data := make([]byte, 500)
	_, err := rd.Read(data)
	if err != nil {
		return err
	}

	message := ssnr.DecodeNotification(data)
	log.Println("Message Received" +
		" from: " + message.GetEmiter() +
		" to:" + string(message.GetReceptor()))
	log.Println("Current list of users: ", users.Length())
	log.Print(users)
	return nil
}

func handleListing(cn net.Conn, rd *bufio.Reader) error {
	defer logAndClose(cn)
	data := make([]byte, 500)
	_, err := rd.Read(data)
	if err != nil {
		return err
	}

	log.Print("Recived listing request")
	listing := ssnr.DecodeListing(data)
	log.Println(" for ", listing.GetAmount(), " users")
	listing.SetUsers(users)
	answer := listing.Encode()
	cn.Write(answer)
	return nil
}

func handleRegister(cn net.Conn, rd *bufio.Reader) error {
	data := make([]byte, 500)
	_, err := rd.Read(data)
	if err != nil {
		return err
	}

	log.Print("Recived register from: ", cn.RemoteAddr())
	req, err := ssnr.DecodeRegister(data)
	if err != nil {
		return err
	}

	storedIndex, err := users.Add(req.GetReceptor(), ssnr.User{req.GetName(), cn})
	switch {
	case err != nil:
		req.SetReturn(ssnr.RefServerFull)
	case storedIndex == req.GetReceptor():
		req.SetReturn(ssnr.ConnAccepted)
	default:
		req.SetReceptor(storedIndex)
		req.SetReturn(ssnr.ConnNewAddres)
	}
	cn.Write(req.Encode())
	return nil
}

func handleDisconnect(cn net.Conn, rd *bufio.Reader) error {
	log.Println("Recived disconnect from: ", cn.RemoteAddr())
	return nil
}

func handleUnknown(cn net.Conn, rd *bufio.Reader) error {
	defer logAndClose(cn)
	data := make([]byte, 500)
	_, err := rd.Read(data)
	if err != nil {
		return err
	}

	log.Printf("Invalid message received!\nCode: %d\n", data[0])
	cn.Write([]byte{0})
	return nil
}

func logAndClose(cn net.Conn) {
	log.Println("Closing to:", cn.RemoteAddr())
	cn.Close()
}
