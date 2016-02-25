package transit_go

type mapEntry struct {
	key   interface{}
	value interface{}
}

type mapEntries Set

type Emitter interface {
	emit(obj interface{}, asMapKey bool, cache WriteCache) error
	emitNil(asMapKey bool, cache WriteCache) error
	emitString(prefix string, tag string, str string, asMapKey bool, cache WriteCache) error
	emitBoolean(b bool, asMapKey bool, cache WriteCache) error
	emitInteger(i int, asMapKey bool, cache WriteCache) error
	emitDouble(f float64, asMapKey bool, cache WriteCache) error
	emitBinary(bytes []byte, asMapKey bool, cache WriteCache) error
	emitArrayStart(size int) error
	emitArrayEnd() error
	emitMapStart(size int) error
	emitMapEnd() error
	emitActualMap(entries mapEntries, ignored bool, cache WriteCache) error
	prefersStrings() bool
	flushWriter() error
}
