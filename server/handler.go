package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

var clients = make(map[net.Conn]string)

func handleClient(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client %s connected.\n", clientAddr)

	var username string

	conn.Write([]byte("Welcome to the chat server!\n"))

	var currentGroup *Group

	go func() {
		for {
			// 클라이언트 10분 타임아웃 설정
			err := conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
			if err != nil {
				fmt.Printf("[Error] setting read deadline: %s\n", err.Error())
				return
			}
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					fmt.Printf("Client %s did not respond for 10 minutes. Disconnectiong \n", clientAddr)
					conn.Write([]byte("Disconnecting, as You did not respond for 10 minutes. \n"))
					if currentGroup != nil {
						currentGroup = LeaveGroup(currentGroup, conn, username)
					}
					delete(clients, conn)
					return
				} else {
					fmt.Printf("Client %s disconnected.\n", clientAddr)
					if currentGroup != nil {
						currentGroup = LeaveGroup(currentGroup, conn, username)
					}
					delete(clients, conn)
					return
				}
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
				if currentGroup == nil {
					conn.Write([]byte("You are not in any group \n"))
				} else {
					currentGroup = LeaveGroup(currentGroup, conn, username)
				}

			} else if strings.HasPrefix(message, "/all ") {
				allMessage := ExtractString(message, "/all")
				BroadcastMessage(username, allMessage)

			} else if currentGroup != nil {
				go func() {
					currentGroup.Messages <- fmt.Sprintf("%s: %s", username, message)
				}()
			} else {
				conn.Write([]byte("👾Invalid command.👾 Enter '/join <group_name>' to join a group or '/create <group_name>' to create a new group.\n"))
			}
		}
	}()

	// 5분마다 클라이언트에게 5분 마다 시간 정보 전달
	ticker := time.NewTicker(5 * time.Minute)
	defer func() {
		ticker.Stop()
		conn.Close()
		mu.Lock()
		delete(clients, conn) // 클라이언트 연결이 종료되면 해당 클라이언트 정보를 제거
		mu.Unlock()
	}()

	for {
		select {
		case <-ticker.C:
			currentTime := time.Now().Format(time.RFC822)
			message := fmt.Sprintf("🕐 Current time: %s 🕐\n ", currentTime)
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println("[Error] sending time to clinet: ", err.Error())
				return
			}
		}
		ticker.Stop()
		return
	}
}
