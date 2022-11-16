package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
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
			conn.Write([]byte(set(args)))
		case "get":
			conn.Write([]byte(get(args)))
		default:
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}

}

func set(args []Value) string {
	if len(args) > 2 {
		if args[2].String() == "px" {
			expiryStr := args[3].String()
			expiryInMilliseconds, err := strconv.Atoi(expiryStr)
			if err != nil {
				return fmt.Sprintf("-ERR PX value (%s) is not an integer\r\n", expiryStr)
			}
			storage.setWithExpiry(args[0].String(), args[1].String(), time.Duration(expiryInMilliseconds))
		} else {
			return fmt.Sprintf("-ERR unknown option for set: %s\r\n", args[2].String())
		}
	} else {
		storage.Set(args[0].String(), args[1].String())
	}
	return "+OK\r\n"
}

func get(args []Value) string {
	value, found := storage.Get(args[0].String())
	if found {
		return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	} else {
		return "$-1\r\n"
	}
}
