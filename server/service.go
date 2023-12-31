package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

var mu sync.Mutex

type Group struct {
	Name     string
	Members  map[net.Conn]string
	Messages chan string
}

var groups = make(map[string]*Group)

// CreateGroup 그룹채팅방 만들기
func CreateGroup(currentGroup *Group, conn net.Conn, groupName string, username string) *Group {
	mu.Lock()
	defer mu.Unlock()

	if currentGroup != nil {
		conn.Write([]byte(fmt.Sprintf("You are currently in a *%s* group. Leave the current group using '/leave' before creating a new one!\n", currentGroup.Name)))
		return currentGroup
	}

	if _, ok := groups[groupName]; !ok {
		newGroup := &Group{
			Name:     groupName,
			Members:  make(map[net.Conn]string),
			Messages: make(chan string, 100),
		}
		newGroup.Members[conn] = username
		groups[groupName] = newGroup
		currentGroup = newGroup
		conn.Write([]byte(fmt.Sprintf("Created and joined group '%s'\n", groupName)))

		newGroup.Messages <- fmt.Sprintf("%s created and joined the group.\n", username)

		go BroadcastGroupMessages(newGroup)
		return currentGroup
	} else {
		conn.Write([]byte(fmt.Sprintf("Group '%s' already exists. Join it using '/join <group_name>'.\n", groupName)))
		return currentGroup
	}

}

// JoinGroup 그룹채팅방 참여
func JoinGroup(currentGroup *Group, conn net.Conn, groupName string, username string) *Group {
	mu.Lock()
	defer mu.Unlock()

	if currentGroup != nil {
		conn.Write([]byte(fmt.Sprintf("You are currently in a *%s* group. Leave the current group using '/leave' before joining a new one!\n", currentGroup.Name)))
		return currentGroup
	}

	if group, ok := groups[groupName]; ok {
		group.Members[conn] = username
		currentGroup = group
		conn.Write([]byte(fmt.Sprintf("Joined group '%s'\n", groupName)))
		go func() {
			group.Messages <- fmt.Sprintf("%s joined the group.\n", username)
		}()
		return currentGroup
	} else {
		conn.Write([]byte(fmt.Sprintf("Group '%s' does not exist. Create it using '/create <group_name>'.\n", groupName)))
		return currentGroup
	}
}

// LeaveGroup 그룹채팅방 나가기
func LeaveGroup(group *Group, conn net.Conn, username string) *Group {

	mu.Lock()
	delete(group.Members, conn)
	if len(group.Members) == 0 {
		close(group.Messages)
		delete(groups, group.Name)
	}
	conn.Write([]byte("Left the chat room🚪 \n"))
	group.Messages <- fmt.Sprintf("%s has left the chat room👋\n", username)
	mu.Unlock()
	return nil
}

// BroadcastGroupMessages 그룹채팅방에 메세지 보내기
func BroadcastGroupMessages(group *Group) {
	for message := range group.Messages {
		for conn := range group.Members {
			conn.Write([]byte(message + "\n"))
		}
	}
}

// GetGroupList 전체 그룹 채팅방 리스트 조회
func GetGroupList(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	var message string

	if len(groups) == 0 {
		conn.Write([]byte("There are no group chat rooms🥲 \n"))
	} else {
		var groupList strings.Builder

		for groupName, group := range groups {
			groupList.WriteString("* " + groupName + " (Members: ")

			for _, member := range group.Members {
				groupList.WriteString(member + " ")
			}
			groupList.WriteString(")  ")
		}

		message = groupList.String()

		conn.Write([]byte(message + "\n"))
	}
}

// BroadcastMessage 모든 사용자에게 메세지 보내기
func BroadcastMessage(sender, message string) {
	for client, username := range clients {
		if client != nil && username != sender {
			client.Write([]byte(fmt.Sprintf("'%s' to all 🔊  %s\n", sender, message)))
		}
	}
}

func TrackClient(conn net.Conn, username string) {
	clients[conn] = username
}
