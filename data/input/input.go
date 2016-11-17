package input

type Config struct {
	In        []string
	Out       []string
	Transform []Func
	Analyse   Func
	Plot      string
}

type Func struct {
	Name   string
	Params []interface{}
}

type SubPrograms struct {
	Count     int
	List      []string
	Plot      []int
	Transform []int
	Merge     []int
	Analyse   []int
}
