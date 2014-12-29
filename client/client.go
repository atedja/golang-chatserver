package client

import "net"
import "container/list"
import "bytes"
import "testtcp/log"

type Client struct {
	Name       string
	Incoming   chan string
	Outgoing   chan string
	Conn       net.Conn
	Quit       chan bool
	ClientList *list.List
}

func (c *Client) Read(buffer []byte) bool {
	bytes_read, err := c.Conn.Read(buffer)
	if err != nil {
		c.Close()
		log.Log(err)
		return false
	}
	log.Log("Read", bytes_read, "bytes")
	return true
}

func (c *Client) Close() {
	c.Quit <- true
	c.Conn.Close()
	c.RemoveMe()
}

func (c *Client) Equal(other *Client) bool {
	return bytes.Equal([]byte(c.Name), []byte(other.Name)) && c.Conn == other.Conn
}

func (c *Client) RemoveMe() {
	for entry := c.ClientList.Front(); entry != nil; entry = entry.Next() {
		client := entry.Value.(Client)
		if c.Equal(&client) {
			log.Log("RemoveMe: ", c.Name)
			c.ClientList.Remove(entry)
		}
	}
}

func IOHandler(Incoming <-chan string, clientList *list.List) {
	for {
		log.Log("IOHandler: Waiting for input")
		input := <-Incoming
		log.Log("IOHandler: Handling ", input)
		for e := clientList.Front(); e != nil; e = e.Next() {
			c := e.Value.(Client)
			c.Incoming <- input
		}
	}
}

// Read data from the client
func ClientReader(c *Client) {
	buffer := make([]byte, 2048)

	for c.Read(buffer) {
		if bytes.Equal(buffer, []byte("/quit")) {
			c.Close()
			break
		}

		data := string(buffer)
		log.Log("ClientReader received from", c.Name, ":", data)
		send := c.Name + " > " + data
		c.Outgoing <- send
		for i := range buffer {
			buffer[i] = 0x00
		}
	}

	c.Outgoing <- c.Name + " has left chat"
	log.Log("ClientReader stopped for ", c.Name)
}

// Sends data to client
func ClientSender(c *Client) {
	for {
		select {
		case buffer := <-c.Incoming:
			// Read through the buffer until it finds the null-terminating character,
			// and sends the data to the client
			data := string(buffer)
			log.Log("ClientSender sending", "\""+data+"\"", "to", c.Name)
			count := 0
			for i := range buffer {
				if buffer[i] == 0x00 {
					break
				}
				count++
			}
			log.Log("Send size: ", count)
			c.Conn.Write([]byte(buffer)[0:count])
		case <-c.Quit:
			// Client is quitting
			log.Log("Client ", c.Name, " quitting")
			c.Conn.Close()
			break
		}
	}
}

// Handles new incoming clients
func ClientHandler(conn net.Conn, ch chan string, clientList *list.List) {
	buffer := make([]byte, 1024)
	bytesRead, error := conn.Read(buffer)
	if error != nil {
		log.Log("Client connection error:", error)
	}

	name := string(buffer[0 : bytesRead-1])
	newClient := &Client{name, make(chan string), ch, conn, make(chan bool), clientList}

	go ClientSender(newClient)
	go ClientReader(newClient)
	clientList.PushBack(*newClient)
	ch <- string("[server] " + name + " has joined the chat")
}
