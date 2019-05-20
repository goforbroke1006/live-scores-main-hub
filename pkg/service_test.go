package pkg

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/goforbroke1006/live-scores-main-hub/pkg/model"
)

var writeFakeMessagesWG sync.WaitGroup

type testFakeProvider struct {
	msgCounter          uint
	msgTotal            uint
	teamNamesSalt       uint
	fakeScore1          uint
	needGoalEventsCount uint
}

func (tfp *testFakeProvider) ReadMessage() (messageType int, p []byte, err error) {
	if tfp.msgCounter >= tfp.msgTotal {
		return -1, nil, errors.New("limit exceeded")
	}
	defer writeFakeMessagesWG.Done()
	tfp.msgCounter++
	msg := fmt.Sprintf(`{"home":"fake %d","away":"fake %d","scores":{"home":%d,"away":0}}`,
		tfp.teamNamesSalt+1,
		tfp.teamNamesSalt+2,
		tfp.fakeScore1,
	)
	if tfp.needGoalEventsCount > 0 {
		tfp.fakeScore1++
		tfp.needGoalEventsCount--
	}
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

type fakeArchiveClient struct {
	mniMux       sync.Mutex
	matchNextID  uint64
	matchStorage map[string]uint64
}

func (fac fakeArchiveClient) RecognizeParticipant(name string) uint64 {
	panic("implement me")
}

func (fac *fakeArchiveClient) RecognizeLiveMatchEvent(team1, team2 string) uint64 {
	val, ok := fac.matchStorage[team1+team2]
	if ok {
		return val
	}

	fac.mniMux.Lock()
	fac.matchStorage[team1+team2] = fac.matchNextID
	fac.matchNextID++
	fac.mniMux.Unlock()

	return fac.matchStorage[team1+team2]
}

func Test_mainHubService_Start(t *testing.T) {
	fakeProv1ExpectedEventsCount := uint(2)
	fakeProv2ExpectedEventsCount := uint(3)
	fakeProviders := map[string]WebSocketConn{
		"fake-provider-1": &testFakeProvider{msgTotal: 5, teamNamesSalt: 10,
			fakeScore1: 0, needGoalEventsCount: fakeProv1ExpectedEventsCount},
		"fake-provider-2": &testFakeProvider{msgTotal: 5, teamNamesSalt: 20,
			fakeScore1: 0, needGoalEventsCount: fakeProv2ExpectedEventsCount},
	}
	fakeConsumers := []WebSocketConn{
		&testFakeConsumer{},
		&testFakeConsumer{},
	}
	fac := &fakeArchiveClient{matchNextID: 1, matchStorage: map[string]uint64{}}

	svc := mainHubService{
		logger:     log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		providers:  fakeProviders,
		consumers:  fakeConsumers,
		dataStream: make(chan model.Updates, 5),
		archive:    fac,
	}

	writeFakeMessagesWG.Add(5 + 5)
	go func() {
		svc.Start()
	}()
	writeFakeMessagesWG.Wait()
	time.Sleep(250 * time.Millisecond)
	close(svc.dataStream)

	expectedCount := uint(len(fakeProviders)) + fakeProv1ExpectedEventsCount + fakeProv2ExpectedEventsCount

	for _, c := range fakeConsumers {
		count := c.(*testFakeConsumer).receivedMesages
		if count != uint(expectedCount) {
			t.Errorf("wrong received messages count - expect %d, actual %d", expectedCount, count)
		}
	}
}
