package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/interviewparrot/OpenAVStream/pkg/mediaserver"
	"github.com/interviewparrot/OpenAVStream/pkg/mediastream"
	"log"
	"net/http"
)


var upgrader = websocket.Upgrader{} // use default options

func ProcessMessage(msg []byte) {
	log.Println("handle incoming bytes")
	clientMessage := mediaserver.ClientMsg{}
	json.Unmarshal(msg, &clientMessage)
	if mediaserver.IsSessionExist(clientMessage.SessionId) {
		session := mediaserver.SessionStore[clientMessage.SessionId]
		switch cmd := clientMessage.Command; cmd {
		case mediaserver.CMD_ReceiveChunk:
			data, err := base64.StdEncoding.DecodeString(clientMessage.Data)
			log.Println("receiving chunk for sessionID: "+ clientMessage.SessionId + " and session state is: " + session.State)
			if err != nil {
				fmt.Println("error:", err)
				return
			}
			mediastream.ProcessIncomingMsg(session, data)
		}
	}
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	session := mediaserver.CreateNewSession(c)
	// Send the session id to the client
	msg := mediaserver.ServerMsg{mediaserver.CMD_ReceiveSessionId, session.SessionId, session.SessionId}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	c.WriteMessage(1, msgBytes);

	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {

			log.Println("read:", err)
			break
		}
		log.Printf("message type: %s", mt)
		if mt == 2 {
			log.Println("Cannot process binary message right now")
		} else {
			ProcessMessage(message)
		}
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("healthy"))
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/session", sessionHandler)
	http.HandleFunc("/ping", pingHandler)
	log.Fatal(http.ListenAndServe(":"+ mediaserver.GetProperty("openavstream.server.port"), nil))
}
