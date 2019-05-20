package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/goforbroke1006/live-scores-main-hub/pkg"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func GetHandlerMiddleware(svc pkg.MainHubService) (h func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, req *http.Request) {
		ws, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				log.Println(err)
			}
			return
		}

		// TODO: send actual data on open connection

		//for _, upd := range sh.GetAll() {
		//	err := ws.WriteJSON(upd)
		//	if nil != err {
		//		log.Println("Error: can not marshal to JSON")
		//	}
		//}
		//*c = append(*c, ws)

		svc.RegisterConsumer(ws)
	}
}
