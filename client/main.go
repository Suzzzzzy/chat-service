package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	serverAddr = "ws://localhost:8080/ws" // 채팅 서버 주소 (서버 주소에 맞게 수정하세요)
)

type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to the chat server")

	go readMessages(conn)

	for {
		messageText := getUserInput()
		if messageText == "/list" {
			// "/list" 명령어를 입력하면 서버에 채널 목록 요청을 보냅니다.
			listChannels(conn)
			fmt.Printf("리스트 조회")
		} else if strings.HasPrefix(messageText, "/create ") {
			// "/create" 명령어를 사용하여 새 채널을 생성합니다.
			channelName := strings.TrimSpace(strings.TrimPrefix(messageText, "/create "))
			createChannel(conn, channelName)
			fmt.Printf("채널 생성")
		} else if strings.HasPrefix(messageText, "/join ") {
			// "/join" 명령어를 사용하여 이미 존재하는 채널에 참여합니다.
			channelName := strings.TrimSpace(strings.TrimPrefix(messageText, "/join "))
			joinChannel(conn, channelName)
		} else {
			// 메시지를 입력하면 메시지를 서버에 전송합니다.
			sendMessage(conn, messageText)
		}
	}
}

func getUserInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter message: ")
	message, _ := reader.ReadString('\n')
	return strings.TrimSpace(message)
}

func readMessages(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("Error reading message: %v", err)
			return
		}
		fmt.Printf("Received: %s\n", msg)
	}
}

func createChannel(conn *websocket.Conn, channelName string) {
	// 서버에게 새 채널 생성 요청을 보냅니다.
	request := Message{
		Sender:  "YourName",
		Content: "/create " + channelName,
	}
	sendJSONMessage(conn, request)
	// 서버로부터 응답을 기다립니다.
	response := Message{}
	err := conn.ReadJSON(&response)
	if err != nil {
		log.Fatalf("Error reading server response: %v", err)
	}

	// 서버로부터 받은 응답에 따라 처리합니다.
	if response.Sender == "Server" {
		if strings.Contains(response.Content, "created successfully") {
			fmt.Println("Channel created successfully.")
		} else if strings.Contains(response.Content, "same name already exists") {
			fmt.Println("Channel with the same name already exists.")
		} else {
			fmt.Println("Server response: " + response.Content)
		}
	}
}

func listChannels(conn *websocket.Conn) {
	// 서버에게 채널 목록 요청을 보냅니다.
	request := Message{
		Sender:  "YourName",
		Content: "/list",
	}
	sendJSONMessage(conn, request)
}

func joinChannel(conn *websocket.Conn, channelName string) {
	// 서버에게 채널 목록 요청을 보냅니다.
	request := Message{
		Sender:  "YourName",
		Content: "/join " + channelName,
	}
	sendJSONMessage(conn, request)
}

func sendMessage(conn *websocket.Conn, messageText string) {
	message := Message{
		Sender:  "YourName",
		Content: messageText,
	}
	sendJSONMessage(conn, message)
}

func sendJSONMessage(conn *websocket.Conn, message Message) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Error encoding message as JSON: %v", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, messageJSON)
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
}
