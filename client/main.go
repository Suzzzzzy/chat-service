package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Print("Enter server address (e.g., localhost:8080): ")
	serverAddress := readInput()

	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	go func() {
		for {
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println("Server disconnected.")
				os.Exit(0)
			}
			fmt.Print(message)
		}
	}()

	for {
		fmt.Print("Enter command: ")
		command := readInput()

		_, err := conn.Write([]byte(command + "\n"))
		if err != nil {
			fmt.Println("Error sending data:", err.Error())
			os.Exit(1)
		}
	}
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
