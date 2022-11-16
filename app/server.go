package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var storage = NewStorage()

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		value, err := DecodeRESP(bufio.NewReader(conn))
		if err != nil {
			fmt.Println("error decoding RESP:", err.Error())
			return
		}
		command := value.Array()[0].String()
		args := value.Array()[1:]
		switch command {
		case "ping":
			conn.Write([]byte("+PONG\r\n"))
		case "echo":
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[0].String()), args[0].String())))
		case "set":
			storage.Set(args[0].String(), args[1].String())
			conn.Write([]byte("+OK\r\n"))
		case "get":
			conn.Write([]byte(fmt.Sprintf("+%s\r\n", storage.Get(args[0].String()))))
		default:
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}

}
