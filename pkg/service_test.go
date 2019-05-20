package pkg

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/goforbroke1006/live-scores-main-hub/pkg/archive"
	"github.com/goforbroke1006/live-scores-main-hub/pkg/model"
)

type testFakeProvider struct {
	msgCounter            uint
	msgTotal              uint
	msgWG                 *sync.WaitGroup
	teamNamesSalt         uint
	fakeScore1            uint
	needGoalEventsCount   uint
	needGoalEventsCounter uint
}

func (tfp *testFakeProvider) ReadMessage() (messageType int, p []byte, err error) {
	if tfp.msgCounter >= tfp.msgTotal {
		return -1, nil, errors.New("limit exceeded")
	}
	defer tfp.msgWG.Done()
	tfp.msgCounter++
	msg := fmt.Sprintf(`{"home":"fake %d","away":"fake %d","scores":{"home":%d,"away":0}}`,
		tfp.teamNamesSalt+1,
		tfp.teamNamesSalt+2,
		tfp.fakeScore1,
	)
	if tfp.needGoalEventsCounter < tfp.needGoalEventsCount {
		tfp.fakeScore1++
		tfp.needGoalEventsCounter++
	}
	return 0, []byte(msg), nil
}

func (tfp testFakeProvider) WriteJSON(v interface{}) error {
	return nil
}

func (tfp testFakeProvider) Close() error {
	return nil
}

type testFakeConsumer struct {
	receivedMessages uint
}

func (tfc testFakeConsumer) ReadMessage() (messageType int, p []byte, err error) {
	return -1, nil, nil
}

func (tfc *testFakeConsumer) WriteJSON(v interface{}) error {
	tfc.receivedMessages++
	return nil
}

func (tfc testFakeConsumer) Close() error {
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
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	fac := &fakeArchiveClient{matchNextID: 1, matchStorage: map[string]uint64{}}

	type fields struct {
		logger    *log.Logger
		archive   archive.ArchiveService
		providers map[string]WebSocketConn
		consumers []WebSocketConn
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Without providers",
			fields: fields{
				logger:    logger,
				archive:   fac,
				providers: map[string]WebSocketConn{},
				consumers: []WebSocketConn{
					&testFakeConsumer{},
					&testFakeConsumer{},
					&testFakeConsumer{},
					&testFakeConsumer{},
					&testFakeConsumer{},
				},
			},
		},
		{
			name: "Simple case",
			fields: fields{
				logger:  logger,
				archive: fac,
				providers: map[string]WebSocketConn{
					"fake-provider-1": &testFakeProvider{msgTotal: 5, teamNamesSalt: 10,
						fakeScore1: 0, needGoalEventsCount: 2},
					"fake-provider-2": &testFakeProvider{msgTotal: 5, teamNamesSalt: 20,
						fakeScore1: 0, needGoalEventsCount: 4},
				},
				consumers: []WebSocketConn{
					&testFakeConsumer{},
					&testFakeConsumer{},
					&testFakeConsumer{},
					&testFakeConsumer{},
					&testFakeConsumer{},
				},
			},
		},
	}
	for _, tt := range tests {
		var writeFakeMessagesWG sync.WaitGroup
		t.Run(tt.name, func(t *testing.T) {
			svc := &mainHubService{
				logger:     tt.fields.logger,
				archive:    tt.fields.archive,
				providers:  tt.fields.providers,
				consumers:  tt.fields.consumers,
				dataStream: make(chan model.Updates, 10),
			}

			mgsCountWait := uint(0)
			for _, p := range tt.fields.providers {
				p.(*testFakeProvider).msgWG = &writeFakeMessagesWG
				mgsCountWait += p.(*testFakeProvider).msgTotal
			}

			writeFakeMessagesWG.Add(int(mgsCountWait))
			go func() {
				svc.Start()
			}()
			writeFakeMessagesWG.Wait()
			time.Sleep(250 * time.Millisecond)
			close(svc.dataStream)

			expectedCount := uint(len(tt.fields.providers))
			for _, p := range tt.fields.providers {
				tmp := p.(*testFakeProvider).needGoalEventsCount
				expectedCount += tmp
			}

			for _, c := range tt.fields.consumers {
				count := c.(*testFakeConsumer).receivedMessages
				if count != uint(expectedCount) {
					t.Errorf("wrong received messages count - expect %d, actual %d", expectedCount, count)
				}
			}
		})
	}
}
