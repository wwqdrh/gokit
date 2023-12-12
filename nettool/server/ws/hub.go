package ws

import (
	"sync"
)

// if (window["WebSocket"]) {
// 	conn = new WebSocket("ws://" + document.location.host + "/ws");
// 	conn.onclose = function (evt) {
// 		var item = document.createElement("div");
// 		item.innerHTML = "<b>Connection closed.</b>";
// 		document.body.append(item);
// 	};
// 	conn.onmessage = function (evt) {
// 		var messages = evt.data.split('\n');
// 		for (var i = 0; i < messages.length; i++) {
// 			var item = document.createElement("div");
// 			item.innerText = messages[i];
// 			document.body.append(item);
// 		}
// 	};
// } else {
// 	var item = document.createElement("div");
// 	item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
// 	document.body.append(item);
// }

var hubPool = sync.Map{}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type WSHub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *WSHub {
	return &WSHub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
