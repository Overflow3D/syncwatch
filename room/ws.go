package room

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

//Peer struct for peers
type Peer struct {
	ws          *websocket.Conn
	name        string
	sendQueue   chan []byte
	id          string
	roomID      string
	numMessages int
	lastMessage time.Time
	suspended   bool
}

//NewPeer create new instance of peer
func NewPeer(ws *websocket.Conn, id string, roomID string, numM int, nick string) *Peer {
	return &Peer{ws, nick, make(chan []byte), id, roomID, numM, time.Now(), false}
}

//Listen listens to all incoming ws messages
func (p *Peer) Listen() {
	defer func() {
		if _, is := rooms[p.roomID]; is {
			rooms[p.roomID].unregister <- p
		}
		p.ws.Close()
	}()

	p.ws.SetReadLimit(int64(1500))
	p.ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

	p.ws.SetPongHandler(func(string) error {
		p.ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
		return nil
	})

	for {
		_, m, e := p.ws.ReadMessage()

		if e != nil {
			break
		}
		//Send message with peer info and do action with it before sending it back to all users
		p.preMessages(m)

	}
}

//Talk sends message from websocket connection to room(hub)
func (p *Peer) Talk() {
	ping := time.NewTicker(5 * time.Second)
	//On defer stop pinging and close peer ws
	defer func() {
		ping.Stop()
		p.ws.Close()
	}()

	for {

		select {
		//case new message send message to room
		case m, k := <-p.sendQueue:
			if !k {
				p.write(websocket.CloseMessage, []byte(""))
				return
			}
			if err := p.write(websocket.TextMessage, m); err != nil {
				return
			}
		//case ping send ping
		case <-ping.C:
			if err := p.write(websocket.PingMessage, []byte("")); err != nil {
				return
			}
		}

	}
}

func (p *Peer) preMessages(msg []byte) {
	data := struct {
		Action string `json:"action"`
		Link   string `json:"link,omitempty"`
		Msg    string `json:"msg,omitempty"`
	}{}

	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	switch data.Action {
	case "newVideo":

		//Send contex of message to all users of the room
		if room, exists := rooms[p.roomID]; exists {
			ctx := struct {
				Action string `json:"action"`
				User   string `json:"user"`
				Link   string `json:"link"`
			}{
				data.Action,
				p.name,
				data.Link,
			}

			room.broadcast(marshalContent(ctx))
		}

	case "msg":
		if len(data.Msg) == 0 {
			return
		}
		if room, exists := rooms[p.roomID]; exists {
			ctx := struct {
				Action string `json:"action"`
				User   string `json:"user"`
				Msg    string `json:"msg"`
			}{
				data.Action,
				p.name,
				data.Msg,
			}

			room.broadcast(marshalContent(ctx))
		}

	}
}

//Write writes to peer
func (p *Peer) write(mType int, payload []byte) error {
	if len(payload) != 0 {
		log.Println("Payload:", payload) // for debugging
	}

	p.ws.SetWriteDeadline(time.Now().Add(time.Duration(3 * time.Second)))
	return p.ws.WriteMessage(mType, payload)
}
