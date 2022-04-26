// package server
//
// import (
// 	"fmt"
// 	"net/http"
//
// 	"github.com/gorilla/mux"
// )
//
// type Server struct{}
//
// func (s *Server) Start(addr string) error {
// 	r := s.initRouting()
// 	err := http.ListenAndServe(addr, r)
// 	return err
// }
//
// func (s *Server) Stop() error {
// 	return nil
// }
//
// func (s *Server) initRouting() *mux.Router {
// 	r := mux.NewRouter()
// 	r.HandleFunc("/", s.handleIndex)
// 	return r
// }
//
// func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
// }
//
// func NewServer() *Server {
// 	s := Server{}
// 	return &s
// }

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	// "github.com/marcelbeumer/crispy-octo-goggles/chatbox/base"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
)

type server struct {
	room *room.Room
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func StartServer(addr string) error {
	s := server{}
	s.room = room.NewRoom() // FIXME: never gets closed
	r := mux.NewRouter()
	r.Handle("/{uuid}", &s)
	return http.ListenAndServe(addr, r)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	log.Printf("Request from %s (uuid \"%s\")\n", r.RemoteAddr, uuid)

	if uuid == "" {
		err := errors.New("Stopping: no uuid")
		writeHttp500(w, err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		err := fmt.Errorf("Could not upgrade connection: %s\n", err)
		writeHttp500(w, err)
		return
	}

	log.Printf("New websocket conn with %s\n", conn.RemoteAddr())

	in, out, err := s.room.Connect(uuid)
	if err != nil {
		err := fmt.Errorf("Could not connect to channel: %s\n", err)
		writeHttp500(w, err)
		return
	}

	_, _ = in, out

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %s\n", err)
			return
		}
		log.Printf("Forwarding message of type %d: %s\n", messageType, p)
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("Write error: %s\n", err)
			return
		}
	}
}

func writeHttp500(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	data := map[string]string{"error": err.Error()}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte(fmt.Sprintf(`Unknown error: %s`, err)))
	}
}
