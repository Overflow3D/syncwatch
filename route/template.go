package route

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-contrib/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/overflow3d/website/room"
)

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

//TODO change roomPage get http response to json response
func roomPage(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)
	ctx := map[string]interface{}{
		"Room":  room,
		"Page":  "room",
		"Title": "Join #" + room.ID,
	}
	fmt.Fprint(w, ctx)
}

//Req request struct
type Req struct {
	Psswd string `json:"psswd"`
}

//Res response struct
type Res struct {
	Code   int    `json:"id"`
	Name   string `json:"name"`
	ErrMsg string `json:"errorMsg,omitempty"`
}

//createRoom creats room and sends Respon
func createRoom(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	//Read request from front-end API
	req := &Req{}
	err := req.decode(r)
	if err != nil {
		return
	}

	//Valids pass and hash it if everything is alright
	res, hash := hashPassword(req.Psswd)
	if res != nil {
		resJSON(w, res.Code, res.Name, res.ErrMsg)
		return
	}

	//Create room Name/URL
	roomID := room.RandString(8)
	_, e := room.InitRoom(roomID, hash)
	if e != nil {
		resJSON(w, 2, "", "Creating room related issue, try again later")
		return

	}

	//Serve resJSOn
	resJSON(w, 0, roomID, "Room created")

}

//login creates cookie if password matches
func login(w http.ResponseWriter, r *http.Request) {
	roomS := r.Context().Value("room").(*room.Room)
	req := &Req{}
	err := req.decode(r)
	if err != nil {
		return
	}

	//Check if password matches, generate uuid
	err = bcrypt.CompareHashAndPassword(roomS.GetPassword(), []byte(req.Psswd))
	if err == nil {
		token := uuid.NewV4()
		resJSON(w, 0, token.String(), "ok")
		return
	}

	resJSON(w, 403, "", "Nie jesteś autoryzowany")

	//return error here

}

//setResponse sets reponse for resJSON function
func setResponse(c int, n string, eMsg string) *Res {
	return &Res{Code: c, Name: n, ErrMsg: eMsg}
}

//Editing it soon
//I love switch please don't judge me :(
func resJSON(w http.ResponseWriter, c int, n string, eMsg string) {
	w.Header().Set("Content-Type", "application/json")
	var res *Res
	switch c {
	case 0:
		w.WriteHeader(http.StatusOK)
		res = setResponse(c, n, eMsg)
		break
	case 1:
		res = setResponse(c, n, eMsg)
		break
	case 2:
		res = setResponse(c, n, eMsg)
		break
	case 403:
		w.WriteHeader(http.StatusForbidden)
		res = setResponse(c, n, eMsg)
		break
	default:
		break
	}

	json.NewEncoder(w).Encode(res)

}

//hashPassword, valids and hash password
func hashPassword(psswd string) (*Res, []byte) {
	r := new(Res)
	//Password lenght validation
	if len(psswd) < 6 || len(psswd) > 10 {
		r = setResponse(1, "", "Password is to short or to long, try again")
		return r, nil
	}

	//Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(psswd), 5)
	if err != nil {
		r = setResponse(1, "", "Password couldn't be generated, try again")
		return r, nil
	}

	return nil, hash

}

//Custome decoder for requests
func (req *Req) decode(r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(req)
	if err != nil {
		return err
	}
	return nil
}
