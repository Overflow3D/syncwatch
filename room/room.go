package room

import (
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
		counter:        1,
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
			if ok {
				log.Println("Peer registered")
				r.peers[peer] = true
			} else {
				break loop
			}
		case peer, ok := <-r.unregister:
			{
				if ok {
					r.unRegisterPeer(peer)
				}
			}

		case m, ok := <-r.broadcastQueue:
			if ok {
				log.Println("Message", string(m))
				for peer := range r.peers {
					log.Println(peer.id)
					select {
					case peer.sendQueue <- m:
						log.Println("Message sended to: ", peer.id)
					}
				}
			}
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
}

//unRegisterPeer unregisters peer from room
func (r *Room) unRegisterPeer(p *Peer) {
	delete(r.peers, p)
	log.Println("Unregistered from room:", p.name, p.id)
}

//GetRoomPeers returns ammount of peers on room
func (r *Room) GetRoomPeers() int {
	return len(r.peers)
}

func (r *Room) broadcast(payload []byte) {
	if len(r.peers) == 0 {
		return
	}

	r.broadcastQueue <- payload

}

func (r *Room) userList() map[string]string {
	list := make(map[string]string)
	return list
}
