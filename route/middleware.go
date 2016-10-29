package route

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/overflow3d/website/room"
)

//Wrapper for httprouter to be more friendly to middleware
func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), "roomID", ps.ByName("room"))
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

//Checks if room exists in map of rooms
func (env *Env) checkIfRoomExists(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Get roomid from request
		roomID := r.Context().Value("roomID")
		str, ok := roomID.(string)
		if !ok {
			errorHandler(w, r, http.StatusNotFound)
			return
		}
		//Check if room is active in memorie
		roomToJoin := room.GetRoom(str)
		if roomToJoin == nil {
			//Handles error if there is no such room as roomID in db nor memorie
			errorHandler(w, r, http.StatusNotFound)
			return
		}

		//Add room to context and chain it
		ctx := context.WithValue(r.Context(), "room", roomToJoin)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)

	})
}

func isAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		h.ServeHTTP(w, r)

	})
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "custom 404")
	}
}
