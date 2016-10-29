package route

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/overflow3d/website/room"
	"github.com/rs/cors"
)

// Server informations
// TODO server config file
type Server struct {
	Port string
}

//Env structure hold db interface and other middleware structures
type Env struct {
	// No need database for now so commented
	// db dbstorage.DataStore
}

var (
	// Websocket http upgrader
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// Run starts the server
func Run(httpHandlers http.Handler) {
	server := Server{":8080"}
	runHTTP(httpHandlers, server)
}

func runHTTP(handlers http.Handler, s Server) {
	log.Println("Starting local server on port", s.Port)

	log.Fatal(http.ListenAndServe(s.Port, handlers))
}

////////////////   ROUTES  ///////////////////

//DoRoutes creates routes and supply them with middleware
func DoRoutes() http.Handler {
	env := &Env{}
	r := httprouter.New()
	r.GET("/", index)
	r.GET("/room/:room", wrapHandler(alice.New(env.checkIfRoomExists).ThenFunc(roomPage)))
	r.GET("/ws/:room", wrapHandler(alice.New(env.checkIfRoomExists, isAuth).ThenFunc(wshandler)))

	r.POST("/create", createRoom)
	r.POST("/room/:room/login", wrapHandler(alice.New(env.checkIfRoomExists).ThenFunc(login)))
	return toHandler(r)
}

//Router to Handler
func toHandler(r *httprouter.Router) http.Handler {
	// Cors for cross-origin resources sharing needed
	// In dev phase ng2 :3000 port and api 8080
	return http.Handler(cors.Default().Handler(r))
}

//////////////// WebSocket Handler ////////////////

func wshandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Println("Wrong")
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrader error", err)
	}
	roomID := r.Context().Value("roomID")
	str := roomID.(string)
	nick := r.URL.Query().Get("nick")

	//Create and register peer to room
	uID := room.RandString(8)
	peer := room.NewPeer(conn, uID, str, 1, nick)
	room := r.Context().Value("room").(*room.Room)
	room.RegisterPeer(peer)

	//Start broadcaster
	go peer.Talk()

	//Start listner
	go peer.Listen()

}
