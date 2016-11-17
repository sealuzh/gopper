package input

import (
	"fmt"
	"reflect"
)

func checkParams(name string, f Func, pos int) error {
	pc := len(f.Params)
	if pc == 0 {
		return fmt.Errorf("%s: no parameters provided", name)
	}

	if pos >= pc || pos < 0 {
		return fmt.Errorf("%s: index (%d) out of bounds. Must be between 0 and %d", name, pos, pc-1)
	}
	return nil
}

func StringParam(f Func, pos int) (string, error) {
	err := checkParams("StringParam", f, pos)
	if err != nil {
		return "", err
	}

	p := f.Params[pos]
	switch p := p.(type) {
	case string:
		return p, nil
	default:
		return "", fmt.Errorf("stringParam: parameter at position %d is not a string (is %v)", pos, reflect.TypeOf(p))
	}
}

func Float32Param(f Func, pos int) (float32, error) {
	err := checkParams("Float32Param", f, pos)
	if err != nil {
		return 0, err
	}

	p := f.Params[pos]
	switch p := p.(type) {
	case float64:
		return float32(p), nil
	default:
		return 0, fmt.Errorf("%s parameter is of incompatible type: %v", f.Name, reflect.TypeOf(p))
	}
}

func IntParam(f Func, pos int) (int, error) {
	err := checkParams("IntParam", f, pos)
	if err != nil {
		return 0, err
	}

	p := f.Params[pos]
	switch p := p.(type) {
	case float64:
		return int(p), nil
	default:
		return 0, fmt.Errorf("%s parameter is of incompatible type: %v", f.Name, reflect.TypeOf(p))
	}
}
