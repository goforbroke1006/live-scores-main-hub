package pkg

import (
	"errors"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/goforbroke1006/live-scores-main-hub/pkg/model"
)

var writeFakeMessagesWG sync.WaitGroup

type testFakeProvider struct {
	readedMgsCount uint
	countMesages   uint
}

func (tfp *testFakeProvider) ReadMessage() (messageType int, p []byte, err error) {
	if tfp.readedMgsCount >= tfp.countMesages {
		return -1, nil, errors.New("limit exceeded")
	}
	defer writeFakeMessagesWG.Done()
	tfp.readedMgsCount++
	msg := `{"home":"fake 1","away":"fake 2","scores":{"home":1,"away":0}}`
	return 0, []byte(msg), nil
}

func (tfp testFakeProvider) WriteJSON(v interface{}) error {
	//panic("implement me")
	return nil
}

func (tfp testFakeProvider) Close() error {
	//panic("implement me")
	return nil
}

type testFakeConsumer struct {
	receivedMesages uint
}

func (tfc testFakeConsumer) ReadMessage() (messageType int, p []byte, err error) {
	return -1, nil, nil
}

func (tfc *testFakeConsumer) WriteJSON(v interface{}) error {
	//panic("implement me")
	tfc.receivedMesages++
	return nil
}

func (tfc testFakeConsumer) Close() error {
	//panic("implement me")
	return nil
}

func Test_mainHubService_Start(t *testing.T) {
	fakeProviders := map[string]WebSocketConn{
		"fake-provider-1": &testFakeProvider{countMesages: 5},
		"fake-provider-2": &testFakeProvider{countMesages: 5},
	}
	fakeConsumers := []WebSocketConn{
		&testFakeConsumer{},
		&testFakeConsumer{},
	}

	svc := mainHubService{
		logger:     log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		providers:  fakeProviders,
		consumers:  fakeConsumers,
		dataStream: make(chan model.Updates, 5),
	}

	writeFakeMessagesWG.Add(5 + 5)
	go func() {
		svc.Start()
	}()
	writeFakeMessagesWG.Wait()
	time.Sleep(250 * time.Millisecond)
	close(svc.dataStream)

	for _, c := range fakeConsumers {
		count := c.(*testFakeConsumer).receivedMesages
		if count != 10 {
			t.Errorf("wrong received messages count - expect %d, actual %d", 10, count)
		}
	}
}
