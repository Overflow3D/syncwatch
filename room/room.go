package room

import (
	"encoding/json"
	"log"
	"time"
)

//Room contains rooms structure
type Room struct {
	ID        string
	psswd     []byte
	timestamp time.Time
	sessions  []string

	peers          map[*Peer]bool
	broadcastQueue chan []byte
	register       chan *Peer
	unregister     chan *Peer
	stop           chan int
	counter        int
}

//Roomer is an interface that contains all methods for room package
type Roomer interface {
}

//Rooms is memorie
var rooms = make(map[string]*Room)

//GetActiveRooms  retrive map of current active rooms
func GetActiveRooms() map[string]*Room {
	return rooms
}

//GetPassword returns password of room
func (r *Room) GetPassword() []byte {
	return r.psswd
}

//GetRoom checks if room exists and retrives it
func GetRoom(id string) *Room {
	room, exists := rooms[id]
	if exists {
		return room
	}
	return nil
}

//AddTestRoom for dem test purpuse
func AddTestRoom(room string) {
	rooms[room] = &Room{ID: "test"}
}

func addRoomToMemory(id string, p []byte) {

	rooms[id] = &Room{
		ID:             id,
		psswd:          p,
		timestamp:      time.Now(),
		sessions:       []string{},
		peers:          make(map[*Peer]bool),
		broadcastQueue: make(chan []byte),
		register:       make(chan *Peer),
		unregister:     make(chan *Peer),
		stop:           make(chan int),
		counter:        0,
	}

}

//InitRoom initialize room into memorie
func InitRoom(id string, psswd []byte) (*Room, error) {
	room := GetRoom(id)
	if room != nil {
		return room, nil
	}
	addRoomToMemory(id, psswd)
	cRoom := GetRoom(id)
	log.Println("Room initalized: ", cRoom)
	go cRoom.run()

	return cRoom, nil
}

func (r *Room) run() {
	log.Println("Start room", r.ID)
	defer log.Println("Stop room", r.ID)
loop:
	for {
		select {
		case peer, ok := <-r.register:
			log.Println("test")
			if ok {
				r.peers[peer] = true

				//We need to start it in new gorouting so select statment doesn't get blocked
				//Otherwise it get stuck beacuse of we are still in case <-r.register, while
				//case <-r.broadcast need to be done to leave <-r.register
				go r.broadcast(r.doUserList(peer, true)) // notify websocket about user joining
				log.Println("User: ", peer.name, " joins ", r.ID, " room.")

			} else {
				break loop
			}
		case peer, ok := <-r.unregister:
			{
				if ok {

					r.unRegisterPeer(peer)
					//We need to start it in new gorouting so select statment doesn't get blocked
					//Otherwise it get stuck beacuse of we are still in case <-r.unregister, while
					//case <-r.broadcast need to be done to leave <-r.unregister
					go r.broadcast(r.doUserList(peer, false)) // notify websocket about user leaving
					log.Println("User: ", peer.name, " leaves ", r.ID, " room.")
				}
			}

		//Emits message to all users in room
		case m, ok := <-r.broadcastQueue:
			if ok {
				for peer := range r.peers {
					select {
					case peer.sendQueue <- m:
					default:
						close(peer.sendQueue)
						delete(r.peers, peer)
					}
				}
			}
		//Deletes room after 30min of inactivity with out any user in it
		case <-time.After(time.Second * 1800):
			if len(r.peers) == 0 {
				break loop
			}

			continue loop
		}

	}
}

//RegisterPeer registers new peer connection from user to room
func (r *Room) RegisterPeer(p *Peer) {
	r.addPeer()
	r.register <- p
}

func (r *Room) addPeer() {
	r.counter++
	log.Println(r.counter)
}

//unRegisterPeer unregisters peer from room
func (r *Room) unRegisterPeer(p *Peer) {
	delete(r.peers, p)
	r.counter--
	log.Println("Unregistered from room:", p.name, p.id)
}

//GetRoomPeers returns ammount of peers on room
func (r *Room) GetRoomPeers() int {
	return len(r.peers)
}

func (r *Room) broadcast(payload []byte) {
	if len(r.peers) > 0 {
		r.broadcastQueue <- payload
	}

}

//doUserList, creates action for websocket with userList and userUpdate status
func (r *Room) doUserList(p *Peer, joining bool) []byte {
	var peerUpdate interface{}
	//Check if peer has information
	if p != nil {
		peerUpdate = struct {
			Peer      string `json:"peer"`
			Name      string `json:"name"`
			IsJoining bool   `json:"joining"`
			Time      int64  `json:"time"`
		}{

			p.id,
			p.name,
			joining,
			time.Now().UTC().Unix(),
		}

	}

	peersInfo := struct {
		Action   string            `json:"action"`
		Message  map[string]string `json:"userList"`
		PeerInfo interface{}       `json:"userInfo"`
	}{
		"userUpdate",
		r.userList(),
		peerUpdate,
	}

	return marshalContent(peersInfo)
}

//userList, retrives userList of specific room
func (r *Room) userList() map[string]string {
	list := make(map[string]string)

	for user := range r.peers {
		list[user.id] = user.name
	}

	return list
}

//marshalContent, marshals interface content into []byte so websocket can handle the message
func marshalContent(content interface{}) []byte {
	if encoded, err := json.Marshal(content); err == nil {
		return encoded
	}
	return []byte("")
}
