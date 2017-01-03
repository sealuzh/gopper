package analyse

import (
	"fmt"
	"reflect"

	"github.com/sealuzh/gopper/data"
	"github.com/senseyeio/roger"
)

const (
	rHost        = "127.0.0.1"
	rPort        = 6311
	rvarTestData = "td"
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

func (rm *rManager) evaluate(tr data.TestResult, stmt string, params ...rParam) (interface{}, error) {
	rc := rm.client()
	s, err := rc.GetSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	d := vectoriseFirstElement(tr)
	err = s.Assign(rvarTestData, d)
	if err != nil {
		return nil, fmt.Errorf("RManager - could not assign test data: %v", err)
	}

	err = assignVariables(s, params...)
	if err != nil {
		return nil, err
	}

	res, err := s.Eval(stmt)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type rParam struct {
	name  string
	value interface{}
}

func assignVariables(s roger.Session, params ...rParam) error {
	for _, param := range params {
		var err error
		switch v := param.value.(type) {
		case string:
			err = s.Assign(param.name, v)
		case []string:
			err = s.Assign(param.name, v)
		case []byte:
			err = s.Assign(param.name, v)
		case []int32:
			err = s.Assign(param.name, v)
		case []float64:
			err = s.Assign(param.name, v)
		default:
			return fmt.Errorf("RManager - unsupported paramater type %v", reflect.TypeOf(v))
		}
		if err != nil {
			return fmt.Errorf("RManager - could not assign parameter '%s = %v': %v", param.name, param.value, err)
		}
	}
	return nil
}
