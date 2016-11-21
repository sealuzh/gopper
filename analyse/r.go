package analyse

import (
	"fmt"

	"github.com/senseyeio/roger"
)

const (
	rHost = "127.0.0.1"
	rPort = 6311
)

type rManager struct {
	host string
	port int64
	c    roger.RClient
}

func newRManager(host string, port int64) *rManager {
	c, err := roger.NewRClient(host, port)
	if err != nil {
		panic(fmt.Sprintf("Could not create RClient for host=%s and port=%d: %v", host, port, err))
	}
	return &rManager{
		host: host,
		port: port,
		c:    c,
	}
}

func newLocalRManager() *rManager {
	return newRManager(rHost, rPort)
}

func (rm *rManager) client() roger.RClient {
	return rm.c
}
