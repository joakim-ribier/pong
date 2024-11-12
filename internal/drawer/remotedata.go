package drawer

import (
	"sort"
	"time"
)

type remoteData struct {
	clients  map[string]*remoteClient
	messages []remoteMessage
}

func newRemoteData() remoteData {
	return remoteData{clients: make(map[string]*remoteClient)}
}

func (r remoteData) clientsSorted() []string {
	keys := make([]string, 0, len(r.clients))
	for k := range r.clients {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type remoteClient struct {
	lastPing time.Time
	lastPong time.Time
}

func newRemoteClient() *remoteClient {
	return &remoteClient{}
}

func (r remoteClient) ping() time.Duration {
	if r.lastPong.After(r.lastPing) {
		return r.lastPong.Sub(r.lastPing)
	}
	return 0 * time.Millisecond
}

type remoteMessage struct {
	dateTime string
	text     string
	level    remoteMessageLevel
}

type remoteMessageLevel int

const (
	info remoteMessageLevel = iota
	warning
)
