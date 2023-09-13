package main

import (
	"fmt"
	"net"
)

type Group struct {
	Name     string
	Members  map[net.Conn]struct{}
	Messages chan string
}

var groups = make(map[string]*Group)

// 그룹채팅방 만들기
func createGroup(currentGroup *Group, conn net.Conn, groupName string, username string) *Group {
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
			newGroup.Messages <- fmt.Sprintf("%s created and joined the group.\n", username)
		}()
		go broadcastGroupMessages(newGroup)
		return currentGroup
	} else {
		mu.Unlock()
		conn.Write([]byte(fmt.Sprintf("Group '%s' already exists. Join it using '/join <group_name>'.\n", groupName)))
		return currentGroup
	}
}

// 그룹채팅방 참여
func joinGroup(currentGroup *Group, conn net.Conn, groupName string, username string) *Group {
	mu.Lock()
	if group, ok := groups[groupName]; ok {
		group.Members[conn] = struct{}{}
		mu.Unlock()
		conn.Write([]byte(fmt.Sprintf("Joined group '%s'\n", groupName)))
		go func() {
			group.Messages <- fmt.Sprintf("%s joined the group.\n", username)
		}()
		return currentGroup
	} else {
		mu.Unlock()
		conn.Write([]byte(fmt.Sprintf("Group '%s' does not exist. Create it using '/create <group_name>'.\n", groupName)))
		return currentGroup
	}
}

// 그룹채팅방 나가기
func leaveGroup(group *Group, conn net.Conn) {
	mu.Lock()
	delete(group.Members, conn)
	if len(group.Members) == 0 {
		close(group.Messages)
		delete(groups, group.Name)
	}
	mu.Unlock()
}

// 그룹채팅방에 메세지 보내기
func broadcastGroupMessages(group *Group) {
	for message := range group.Messages {
		for conn := range group.Members {
			conn.Write([]byte(message + "\n"))
		}
	}
}

// 전체 그룹 채팅방 리스트 조회
func getGroupList(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	if len(groups) == 0 {
		conn.Write([]byte("There are no group chat rooms🥲 \n"))
	} else {
		conn.Write([]byte(fmt.Sprintf("Here is the list of %d available chat rooms🙌 \n", len(groups))))
		for groupName := range groups {
			conn.Write([]byte("* " + groupName + "\n"))
		}
	}
}
