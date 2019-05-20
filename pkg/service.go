package pkg

import (
	"encoding/json"
	"log"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/goforbroke1006/live-scores-main-hub/pkg/model"
)

type WebSocketConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteJSON(v interface{}) error
	Close() error
}

type MainHubService interface {
	RegisterProvider(name, url string)
	RegisterConsumer(conn WebSocketConn)
	Start()
}

type mainHubService struct {
	logger     *log.Logger
	pMux       sync.Mutex
	providers  map[string]WebSocketConn
	cMux       sync.Mutex
	consumers  []WebSocketConn
	sMux       sync.Mutex
	states     map[uint64]model.Updates
	dataStream chan model.Updates
}

func (svc mainHubService) RegisterProvider(name, urlAddr string) {
	u, _ := url.Parse(urlAddr)
	svc.logger.Printf("connecting to %s\n", u.String())

	provConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		svc.logger.Println("dial:", err)
	}
	svc.pMux.Lock()
	svc.providers[name] = provConn
	svc.pMux.Unlock()
}

func (svc *mainHubService) RegisterConsumer(conn WebSocketConn) {
	svc.cMux.Lock()
	svc.consumers = append(svc.consumers, conn)
	svc.cMux.Unlock()
}

func (svc *mainHubService) Start() {
	svc.logger.Println("Start processing...")

	go func() {
		for name, p := range svc.providers {
			provName := name
			provObj := p
			go func() {
				defer p.Close()
				for {
					_, message, err := provObj.ReadMessage()
					if err != nil {
						svc.logger.Println("read:", err)
						break
					}
					svc.logger.Printf("recv [%s] : %s\n", provName, message)

					var upp model.Updates
					_ = json.Unmarshal(message, &upp)

					svc.dataStream <- upp
				}
			}()
		}
	}()

	for chunk := range svc.dataStream {

		// TODO: check data newest

		for ci, conn := range svc.consumers {
			consIndex := ci
			consConn := conn

			go func() {
				err := consConn.WriteJSON(chunk)
				if nil != err {
					_ = consConn.Close()
					svc.cMux.Lock()
					svc.consumers = append(svc.consumers[:consIndex], svc.consumers[consIndex+1:]...)
					svc.cMux.Unlock()
				}
			}()
		}
	}
}

func NewMainHubService(logger *log.Logger) MainHubService {
	return &mainHubService{
		logger:     logger,
		providers:  map[string]WebSocketConn{},
		consumers:  []WebSocketConn{},
		states:     map[uint64]model.Updates{},
		dataStream: make(chan model.Updates, 10),
	}
}
