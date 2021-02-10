package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type Channel struct {
	name      string
	connectCh chan Client
	msgCh     chan Msg
	connected []Client
}

type Client struct {
	name   string
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

type Msg struct {
	sender string
	body   string
}

var channels map[string]Channel

func init() {
	channels = make(map[string]Channel)
	channels["oop"] = newChannel("oop")
	channels["bananas"] = newChannel("bananas")
}

func newChannel(name string) Channel {
	channel := Channel{}
	channel.name = name
	channel.connectCh = make(chan Client)
	channel.msgCh = make(chan Msg)

	go func() {
		for {
			select {
			case client := <-channel.connectCh:
				fmt.Println("User " + client.name + " connected to channel " + channel.name)
				msg := "join " + client.name + "\n"
				for _, who := range channel.connected {
					who.writer.WriteString(msg)
					who.writer.Flush()
				}
				channel.connected = append(channel.connected, client)
			case msg := <-channel.msgCh:
				for _, who := range channel.connected {
					fmt.Println("Sending to " + who.name)
					msgMsg := "'" + msg.sender + "'@'" + channel.name + "'" + msg.body + "\n"
					who.writer.WriteString(msgMsg)
					who.writer.Flush()
				}
			}
		}
	}()

	return channel
}

func main() {
	ln, err := net.Listen("tcp", ":0")
	defer ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	port := ln.Addr().String()[5:]
	fmt.Println("Running at port: " + port)

	for {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			log.Print(err)
		}

		client := newClient(conn)

		go handleClient(client)
	}
}

func (c *Client) Read(p []byte) (int, error) {
	n, err := c.conn.Read(p)
	return n, err
}

func (c *Client) Write(p []byte) (int, error) {
	n, err := c.conn.Write(p)
	return n, err
}

func newClient(conn net.Conn) *Client {
	client := &Client{conn: conn}
	client.reader = bufio.NewReader(client)
	client.writer = bufio.NewWriter(client)
	return client
}

func handleClient(client *Client) {
	msg, _, err := client.reader.ReadLine()
	if err != nil {
		log.Print(err)
		return
	}

	client.name = string(msg)
	if client.name == "weestack" {
		client.name = "peestack"
	}

	channelsMsg := "channels: "
	for name, _ := range channels {
		channelsMsg += "'" + name + "' "
	}
	channelsMsg += "\n"

	_, err = client.writer.WriteString(channelsMsg)
	client.writer.Flush()
	if err != nil {
		log.Print(err)
	}

	// Hardcode oop hehe
	connectToChannel(client, "oop")

	for {
		line, _, err := client.reader.ReadLine()
		if err != nil {
			log.Print(err)
			return
		}
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "tell ") {
			elems := strings.Split(lineStr, "'")
			if len(elems) < 3 {
				continue
			}
			chanName := elems[1]
			msg := strings.Join(elems[2:], "'")
			fmt.Println("channame: " + chanName)
			fmt.Println("msg: " + msg)

			sendToChannel(client, chanName, msg)
		}
	}
}

func connectToChannel(client *Client, name string) {
	channel, ok := channels[name]
	if ok {
		channel.connectCh <- *client
	}
}

func sendToChannel(client *Client, chanName string, msg string) {
	channel, ok := channels[chanName]
	if ok {
		channel.msgCh <- Msg{sender: client.name, body: msg}
	}
}
