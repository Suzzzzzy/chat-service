package main

import (
	"fmt"
	"net"
	"strings"
)

var clients = make(map[net.Conn]string)

func handleClient(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client %s connected.\n", clientAddr)

	var username string

	conn.Write([]byte("Welcome to the chat server!\n"))

	var currentGroup *Group

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Client %s disconnected.\n", clientAddr)
			if currentGroup != nil {
				LeaveGroup(currentGroup, conn)
			}
			delete(clients, conn)
			return
		}

		message := string(buffer[:n])
		message = strings.TrimSpace(message)

		if strings.HasPrefix(message, "/setusername ") {
			username = ExtractString(message, "/setusername")

			conn.Write([]byte(fmt.Sprintf("Your username has been set to: %s\n", username)))

			TrackClient(conn, username)

		} else if strings.HasPrefix(message, "/join ") {
			groupName := ExtractString(message, "/join ")

			currentGroup = JoinGroup(currentGroup, conn, groupName, username)

		} else if strings.HasPrefix(message, "/create ") {
			groupName := ExtractString(message, "/create ")

			currentGroup = CreateGroup(currentGroup, conn, groupName, username)

		} else if strings.HasPrefix(message, "/list") {
			GetGroupList(conn)

		} else if strings.HasPrefix(message, "/leave") {
			currentGroup.Messages <- fmt.Sprintf("%s has left the chat room👋\n", username)
			LeaveGroup(currentGroup, conn)
			currentGroup = nil

		} else if strings.HasPrefix(message, "/all ") {
			allMessage := ExtractString(message, "/all")
			BroadcastMessage(username, allMessage)

		} else if currentGroup != nil {
			go func() {
				currentGroup.Messages <- fmt.Sprintf("%s %s: %s", username, clientAddr, message)
			}()
		} else {
			conn.Write([]byte("👾Invalid command.👾 Enter '/join <group_name>' to join a group or '/create <group_name>' to create a new group.\n"))
		}
	}
}
