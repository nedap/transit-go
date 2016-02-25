package transit_go

type Parser interface {
	parse(cache ReadCache) (interface{}, error)
	parseVal(asMapKey bool, cache ReadCache) (interface{}, error)
	parseMap(asMapKey bool, cache ReadCache, handler *MapReadHandler) (interface{}, error)
	parseArray(asMapKey bool, cache ReadCache, handler *ArrayReadHandler) (interface{}, error)
	parseString(str string) (interface{}, error)
}
