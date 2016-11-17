package analyse

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/senseyeio/roger"
)

const rHost = "127.0.0.1"

var rPort int64 = 6310

type rManager struct {
	host       string
	port       int64
	c          roger.RClient
	createOnce sync.Once
}

func newRManager() *rManager {
	p := atomic.AddInt64(&rPort, 1)
	return &rManager{
		host: rHost,
		port: p,
	}
}

func (m *rManager) client() roger.RClient {
	m.createOnce.Do(func() {
		c, err := roger.NewRClient(m.host, m.port)
		if err != nil {
			panic(fmt.Sprintf("Could not create RClient for host=%s and port=%d", m.host, m.port))
		}
		m.c = c
	})
	return m.c
}
