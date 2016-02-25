package transit_go

type MapReader interface {
	Init() interface{}
	Add(m, key, val interface{}) interface{}
	Complete(interface{}) interface{}
}

type MapBuilder struct{}

func (b MapBuilder) Init() interface{} {
	return make(map[interface{}]interface{})
}

func (b MapBuilder) Add(m interface{}, key, val interface{}) interface{} {
	actualMap, _ := m.(map[interface{}]interface{})
	actualMap[key] = val
	return actualMap
}

func (b MapBuilder) Complete(m interface{}) interface{} {
	return m
}
