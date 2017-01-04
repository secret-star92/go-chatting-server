package lib

import (
	"log"
	"sync"
	"github.com/gorilla/websocket"
	"net/http"
	"net"
)

//////////////////////
//CONNECTIONS
/////////////////////

type connection struct {
	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *Hub
}

func (c *connection) reader(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	defer wg.Done()
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			break
		}
		c.h.broadcast <- message
	}
}

func (c *connection) writer(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	defer wg.Done()
	for message := range c.send {
		err := wsConn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type WsHandler struct {
	H *Hub
}

func (wsh WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//Find users IP and display them
	ip,_,_ := net.SplitHostPort(r.RemoteAddr)
	log.Printf("%s has connected", ip)


	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading %s", err)
		return
	}
	c := &connection{send: make(chan []byte, 256), h: wsh.H}
	c.h.addConnection(c)
	defer c.h.removeConnection(c)
	var wg sync.WaitGroup
	wg.Add(2)
	go c.writer(&wg, wsConn)
	go c.reader(&wg, wsConn)
	wg.Wait()
	wsConn.Close()
}

