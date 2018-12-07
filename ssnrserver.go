package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"strconv"

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
		err = handleNotification(cn, reader)
	case ssnr.ListingCode:
		err = handleListing(cn, reader)
	case ssnr.RegisterCode:
		err = handleRegister(cn, reader)
	default:
		err = handleUnknown(cn, reader)
	}
	if err != nil {
		log.Println(err)
	}
	return nil
}

func handleNotification(cn net.Conn, rd *bufio.Reader) error {
	defer logAndClose(cn)

	data, message, err := ssnr.ReadNotification(rd)
	if err != nil {
		return err
	}
	log.Println("Message Received",
		"from: \""+message.GetEmitter()+"\"",
		"to:", message.GetReceptor())
	usr := users.Get(message.GetReceptor())
	if usr != nil {
		_, err = usr.Addr.Write(data)
		return err
	}
	return errors.New("Message for non indexed user: " +
		strconv.FormatInt(int64(message.GetReceptor()), 10))
}

func handleListing(cn net.Conn, rd *bufio.Reader) error {
	defer logAndClose(cn)
	log.Println("Recived listing request")
	handleDisconnects()
	_, listing, err := ssnr.ReadListing(rd, true)
	if err != nil {
		return err
	}
	listing.SetUsers(users)
	answer := listing.Encode()
	cn.Write(answer)
	return nil
}

func handleRegister(cn net.Conn, rd *bufio.Reader) error {
	_, req, err := ssnr.ReadRegister(rd)
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

func handleDisconnects() {
	log.Println("Cleanning any all connection")
	n := users.CleanDisconnects()
	log.Println(n, "receivers where removed")
}

func logAndClose(cn net.Conn) {
	log.Println("Closing to:", cn.RemoteAddr())
	cn.Close()
}
