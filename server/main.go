package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clients  = make(map[*websocket.Conn]string)
	channels = make(map[string][]*websocket.Conn)
	mutex    = sync.Mutex{}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	username := r.URL.Query().Get("username")
	group := r.URL.Query().Get("group")

	mutex.Lock()
	clients[ws] = username
	channels[group] = append(channels[group], ws)
	mutex.Unlock()

	fmt.Printf("%s joined group %s\n", username, group)

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			removeFromGroup(group, ws)
			break
		}
		msg.Sender = username
		if strings.HasPrefix(msg.Content, "/create ") {
			// "/create" 명령어를 처리합니다.
			channelName := strings.TrimSpace(strings.TrimPrefix(msg.Content, "/create "))
			handleCreateChannel(ws, username, group, channelName)
			fmt.Printf("채널 생성")
		} else if strings.HasPrefix(msg.Content, "/list") {
			handleListChannels(ws)
		} else if strings.HasPrefix(msg.Content, "/join ") {
			// "/join" 명령어를 사용하여 이미 존재하는 채널에 참여합니다.
			channelName := strings.TrimSpace(strings.TrimPrefix(msg.Content, "/join "))
			joinChannel(ws, username, group, channelName)
			fmt.Printf("채널 입장")
		} else {
			// 일반 메시지를 브로드캐스트합니다.
			msg.Sender = username

			broadcast(group, msg)
		}
	}
}

func removeFromGroup(group string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	for i, c := range channels[group] {
		if c == conn {
			channels[group] = append(channels[group][:i], channels[group][i+1:]...)
			break
		}
	}
}

func broadcast(group string, msg Message) {
	mutex.Lock()
	channelClients := channels[group]
	mutex.Unlock()
	for _, client := range channelClients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}
}

func getChannelList() []string {
	mutex.Lock()
	defer mutex.Unlock()

	var channelList []string
	for channelName := range channels {
		channelList = append(channelList, channelName)
	}
	return channelList
}

func handleListChannels(conn *websocket.Conn) {
	// 서버에서 현재 존재하는 채널 목록을 가져옵니다.
	channelList := getChannelList()

	// 채널 목록을 클라이언트에게 전송합니다.
	response := Message{
		Sender:  "Server",
		Content: "Channel List: " + strings.Join(channelList, ", "),
	}
	sendJSONMessage(conn, response)
}
func handleCreateChannel(conn *websocket.Conn, username string, group string, channelName string) {
	// Mutex를 사용하여 동시성 문제를 방지합니다.
	mutex.Lock()
	defer mutex.Unlock()

	// 이미 같은 이름의 채널이 존재하는지 확인합니다.
	if _, exists := channels[channelName]; exists {
		response := Message{
			Sender:  "Server",
			Content: "Channel with the same name already exists.",
		}
		sendJSONMessage(conn, response)
		return
	}

	// 새 채널을 생성합니다.
	channels[channelName] = []*websocket.Conn{conn}

	// 클라이언트에게 성공 메시지를 전송합니다.
	response := Message{
		Sender:  "Server",
		Content: fmt.Sprintf("Channel '%s' created successfully.", channelName),
	}
	sendJSONMessage(conn, response)

	// 사용자를 새 채널에 추가합니다.
	clients[conn] = username
	conn.WriteJSON(Message{
		Sender:  "Server",
		Content: fmt.Sprintf("You joined the channel '%s'.", channelName),
	})

	// 사용자가 채널에 참여했음을 다른 사용자에게 알립니다.
	broadcast(channelName, Message{
		Sender:  "Server",
		Content: fmt.Sprintf("%s joined the channel.", username),
	})
}

func joinChannel(conn *websocket.Conn, username string, group string, channelName string) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := channels[channelName]; !exists {
		response := Message{
			Sender:  "Server",
			Content: fmt.Sprintf("Channel '%s' does not exist.", channelName),
		}
		sendJSONMessage(conn, response)
		return
	}

	channels[channelName] = append(channels[channelName], conn)
	clients[conn] = username
	conn.WriteJSON(Message{
		Sender:  "Server",
		Content: fmt.Sprintf("You joined the channel '%s'.", channelName),
	})

	broadcast(channelName, Message{
		Sender:  "Server",
		Content: fmt.Sprintf("%s joined the channel.", username),
	})
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

func main() {
	http.HandleFunc("/ws", handleConnections)

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	fmt.Println("Group Chat Server started on :8080")
	select {}
}
