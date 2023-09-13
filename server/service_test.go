package main

import (
	"github.com/stretchr/testify/suite"
	"net"
	"testing"
)

type GroupEntityTestSuite struct {
	suite.Suite
}

func (t *GroupEntityTestSuite) SetupTest() {}

func TestGroupEntityTestSuite(t *testing.T) {
	suite.Run(t, new(GroupEntityTestSuite))
}

func (t *GroupEntityTestSuite) TestCreateGroup() {
	t.Run("새로운 채팅방 만들기 - 성공", func() {
		// given
		username := "suji"
		groupName := "New"
		conn := &net.TCPConn{}
		// when
		currentGroup := CreateGroup(nil, conn, groupName, username)
		// then
		t.NotNil(currentGroup)
		t.Equal(groupName, currentGroup.Name)
		t.Equal(1, len(currentGroup.Members)) // 채팅방 생성하자마자 참여되기 때문
		t.Equal(username, currentGroup.Members[conn])
	})
	t.Run("새로운 채팅방 만들기 - 이름 중복으로 실패", func() {
		// given
		username := "suji"
		groupName := "Existing"
		existingGroup := &Group{
			Name:     groupName,
			Members:  make(map[net.Conn]string),
			Messages: make(chan string, 100),
		}
		groups[groupName] = existingGroup

		conn := &net.TCPConn{}

		// when
		currentGroup := CreateGroup(nil, conn, groupName, username)

		// then
		t.Nil(currentGroup)
		t.Equal(existingGroup, groups[groupName])
	})
}

func (t *GroupEntityTestSuite) TestJoinGroup() {
	t.Run("채팅방 참여 - 성공", func() {
		// given
		username := "suji"
		groupName := "Existing"
		existingGroup := &Group{
			Name:     groupName,
			Members:  make(map[net.Conn]string),
			Messages: make(chan string, 100),
		}
		groups[groupName] = existingGroup

		conn := &net.TCPConn{}

		// when
		currentGroup := JoinGroup(nil, conn, groupName, username)

		// then
		t.NotNil(currentGroup)
		t.Equal(groupName, currentGroup.Name)
		t.Equal(1, len(currentGroup.Members))
		t.Equal(username, currentGroup.Members[conn])
	})
	t.Run("채팅방 참여 실패 - 존재하지 않는 채팅방 참여 요청", func() {
		// given
		username := "suji"
		groupName := "NoExisting"
		conn := &net.TCPConn{}

		// when
		currentGroup := JoinGroup(nil, conn, groupName, username)

		// then
		t.Nil(currentGroup)
	})
	t.Run("채팅방 참여 실패 - 현재 참여중인 채팅방이 있을 때", func() {
		// given
		username := "suji"
		groupName := "Existing"
		existingGroup := &Group{
			Name:     groupName,
			Members:  make(map[net.Conn]string),
			Messages: make(chan string, 100),
		}
		groups[groupName] = existingGroup

		conn := &net.TCPConn{}

		// when
		newGroupName := "new"
		currentGroup := JoinGroup(existingGroup, conn, newGroupName, username)

		// then
		t.NotNil(currentGroup)
		t.Equal(groupName, currentGroup.Name)
	})
}
