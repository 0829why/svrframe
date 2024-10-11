package zws

import (
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var (
	upgrader websocket.Upgrader
	cID      atomic.Uint32
)

func checkOrigin(r *http.Request) bool {
	return true
}

func init() {
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     checkOrigin,
	}
	cID.Store(0)
}
