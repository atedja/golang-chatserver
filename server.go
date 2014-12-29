package main

import "net"
import "container/list"
import "testtcp/client"
import "testtcp/log"

func main() {
	log.Log("Hello Server!")
	clientList := list.New()
	in := make(chan string)
	go client.IOHandler(in, clientList)

	service := ":9988"
	tcpAddr, error := net.ResolveTCPAddr("tcp", service)
	if error != nil {
		log.Log("Error: Could not resolve address")
	} else {
		netListen, error := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if error != nil {
			log.Log(error)
		} else {
			defer netListen.Close()

			for {
				log.Log("Waiting for clients")
				connection, error := netListen.Accept()
				if error != nil {
					log.Log("Client error: ", error)
				} else {
					go client.ClientHandler(connection, in, clientList)
				}
			}
		}
	}
}
