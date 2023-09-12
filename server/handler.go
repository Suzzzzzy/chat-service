package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

var groups = make(map[string]*Group)
var mu sync.Mutex

func handleClient(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client %s connected.\n", clientAddr)

	var username string

	conn.Write([]byte("Welcome to the chat server!\n"))
	conn.Write([]byte("Enter '/join <group_name>' to join a group.\n"))

	var currentGroup *Group

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Client %s disconnected.\n", clientAddr)
			if currentGroup != nil {
				leaveGroup(currentGroup, conn)
			}
			return
		}

		message := string(buffer[:n])
		message = strings.TrimSpace(message)

		if strings.HasPrefix(message, "/setusername ") {
			username = trimPrefix(message, "/setusername")

			conn.Write([]byte(fmt.Sprintf("Your username has been set to: %s\n", username)))

		} else if strings.HasPrefix(message, "/join ") {
			groupName := trimPrefix(message, "/join ")

			joinGroup(currentGroup, conn, groupName, username)

		} else if strings.HasPrefix(message, "/create ") {
			groupName := trimPrefix(message, "/create ")

			createGroup(currentGroup, conn, groupName, username)

		} else if strings.HasPrefix(message, "/list") {
			getGroupList(conn)

		} else if currentGroup != nil {
			go func() {
				currentGroup.Messages <- fmt.Sprintf("%s %s: %s", username, clientAddr, message)
			}()
		} else {
			conn.Write([]byte("👾Invalid command.👾 Enter '/join <group_name>' to join a group or '/create <group_name>' to create a new group.\n"))
		}
	}
}
