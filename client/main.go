package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	username string
	port     int
)

func init() {
	flag.StringVar(&username, "username", "", "Your username")
	flag.IntVar(&port, "port", 8080, "Port to connect to")
}

func main() {
	// 사용자 이름과 포트 번호를 입력받음
	flag.Parse()

	if username == "" {
		fmt.Println("Please provide a username using the -username flag.")
		os.Exit(1)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "/setusername %s\n", username)

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
		fmt.Print("> ")
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
