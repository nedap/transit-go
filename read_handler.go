package transit_go

type ReadHandler struct {
	Name    string
	FromRep func(rep interface{}) (interface{}, error)
}

type DefaultReadHandler struct {
	FromRep func(tag string, rep interface{}) (interface{}, error)
}

type ArrayReadHandler struct {
	ReadHandler
	arrayReader ArrayReader
}

type MapReadHandler struct {
	ReadHandler
	mapReader MapReader
}
