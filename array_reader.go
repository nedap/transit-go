package transit_go

type ArrayReader interface {
	Init(size int) interface{}
	Add(a interface{}, item interface{}) interface{}
	Complete(a interface{}) interface{}
}

type ArrayBuilder struct{}

func (b ArrayBuilder) Init(size int) interface{} {
	return make([]interface{}, size, size)
}

func (b ArrayBuilder) Add(a interface{}, item interface{}) interface{} {
	actualList, _ := a.([]interface{})
	res := append(actualList, item)
	return res
}

func (b ArrayBuilder) Complete(a interface{}) interface{} {
	return a
}
