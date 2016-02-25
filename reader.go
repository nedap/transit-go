package transit_go

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type TransmitReader interface {
	Read() interface{}
}

type ReadHandlerMap map[string]interface{}

type transmitReader struct {
	parser   Parser
	handlers ReadHandlerMap
}

type JSONReader struct {
	transmitReader
}

func defaultReadHandlers() ReadHandlerMap {
	handlers := ReadHandlerMap{
		":":     keywordReadHandler(),
		"$":     symbolReadHandler(),
		"i":     integerReadHandler(),
		"?":     booleanReadHandler(),
		"_":     nullReadHandler(),
		"f":     bigDecimalReadHandler(),
		"n":     bigIntegerReadHandler(),
		"d":     doubleReadHandler(),
		"z":     specialNumberReadHandler(),
		"c":     characterReadHandler(),
		"t":     timeReadHandler(),
		"m":     timeReadHandler(),
		"r":     uriReadHandler(),
		"u":     uuidReadHandler(),
		"b":     binaryReadHandler(),
		"'":     identityReadHandler(),
		"set":   setReadHandler(),
		"list":  listReadHandler(),
		"ratio": ratioReadHandler(),
		"cmap":  cmapReadHandler(),
		"link":  linkReadHandler(),
	}
	return handlers
}

func NewJSONReader(buffer *bytes.Buffer) JSONReader {
	return NewJSONReaderWithHandlers(buffer, map[string]ReadHandler{})
}

func NewJSONReaderWithHandlers(buffer *bytes.Buffer, customHandlers map[string]ReadHandler) JSONReader {
	handlers := defaultReadHandlers()

	for tag, handler := range customHandlers {
		handlers[tag] = handler
	}
	jsonDecoder := json.NewDecoder(buffer)
	jsonDecoder.UseNumber()

	parser := NewJsonParser(jsonDecoder, handlers, defaultReadHandler(), defaultMapBuilder(), defaultListBuilder())
	reader := JSONReader{
		transmitReader{
			handlers: handlers,
			parser:   parser,
		},
	}
	return reader
}

func defaultReadHandler() *DefaultReadHandler {
	return &DefaultReadHandler{
		FromRep: func(tag string, rep interface{}) (interface{}, error) {
			return TaggedValue{Tag: tag, Rep: rep}, nil
		},
	}
}

func defaultMapBuilder() MapBuilder {
	return MapBuilder{}
}

func defaultListBuilder() ArrayBuilder {
	return ArrayBuilder{}
}

func (m ReadHandlerMap) lookupHandler(tag string) (interface{}, error) {
	handler, found := m[tag]
	if !found {
		return ReadHandler{}, fmt.Errorf("No handler found for tag %s", tag)
	}
	return handler, nil
}

func (r JSONReader) Read() interface{} {
	val, err := r.parser.parse(NewReadCache())
	if err != nil {
		panic(err)
	}
	return val
}
