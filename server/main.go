package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Group struct {
	Name     string
	Members  map[net.Conn]struct{}
	Messages chan string
}

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
			username = strings.TrimPrefix(message, "/setusername")
			username = strings.TrimSpace(username)
			conn.Write([]byte(fmt.Sprintf("Your username has been set to: %s\n", username)))
		} else if strings.HasPrefix(message, "/join ") {
			groupName := strings.TrimPrefix(message, "/join ")
			groupName = strings.TrimSpace(groupName)

			mu.Lock()
			if group, ok := groups[groupName]; ok {
				group.Members[conn] = struct{}{}
				currentGroup = group
				mu.Unlock()
				conn.Write([]byte(fmt.Sprintf("Joined group '%s'\n", groupName)))
				go func() {
					group.Messages <- fmt.Sprintf("%s %s joined the group.\n", username, clientAddr)
				}()
			} else {
				mu.Unlock()
				conn.Write([]byte(fmt.Sprintf("Group '%s' does not exist. Create it using '/create <group_name>'.\n", groupName)))
			}
		} else if strings.HasPrefix(message, "/create ") {
			groupName := strings.TrimPrefix(message, "/create ")
			groupName = strings.TrimSpace(groupName)

			mu.Lock()
			if _, ok := groups[groupName]; !ok {
				newGroup := &Group{
					Name:     groupName,
					Members:  make(map[net.Conn]struct{}),
					Messages: make(chan string, 100),
				}
				newGroup.Members[conn] = struct{}{}
				groups[groupName] = newGroup
				currentGroup = newGroup
				mu.Unlock()
				conn.Write([]byte(fmt.Sprintf("Created and joined group '%s'\n", groupName)))
				go func() {
					newGroup.Messages <- fmt.Sprintf("Client %s created and joined the group.\n", clientAddr)
				}()
				go broadcastGroupMessages(newGroup)
			} else {
				mu.Unlock()
				conn.Write([]byte(fmt.Sprintf("Group '%s' already exists. Join it using '/join <group_name>'.\n", groupName)))
			}
		} else if strings.HasPrefix(message, "/list") {
			getGroupList(conn)
		} else if currentGroup != nil {
			go func() {
				currentGroup.Messages <- fmt.Sprintf("%s %s: %s", username, clientAddr, message)
			}()
		} else {
			conn.Write([]byte("Invalid command. Enter '/join <group_name>' to join a group or '/create <group_name>' to create a new group.\n"))
		}
	}
}

func leaveGroup(group *Group, conn net.Conn) {
	mu.Lock()
	delete(group.Members, conn)
	if len(group.Members) == 0 {
		close(group.Messages)
		delete(groups, group.Name)
	}
	mu.Unlock()
}

func broadcastGroupMessages(group *Group) {
	for message := range group.Messages {
		for conn := range group.Members {
			conn.Write([]byte(message + "\n"))
		}
	}
}

func getGroupList(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock() // mutex 해제

	if len(groups) == 0 {
		conn.Write([]byte("There are no group chat rooms🥲 \n"))
	} else {
		conn.Write([]byte("Here is the list of available chat rooms🙌 \n"))
		for groupName := range groups {
			conn.Write([]byte("* " + groupName + "\n"))
		}
	}
}

func main() {
	listen, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listen.Close()
	fmt.Println("Chat server is listening on localhost:8080")

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			os.Exit(1)
		}
		go handleClient(conn)
	}
}
