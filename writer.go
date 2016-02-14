package transit_go

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"

	"github.com/twinj/uuid"
)

type TransmitWriter interface {
	Write(interface{}) error
}

type WriteHandler struct {
	Tag       string
	Rep       func(interface{}) interface{}
	StringRep func(interface{}) interface{}
}

type writeHandlers map[reflect.Type]WriteHandler

type transmitWriter struct {
	buffer   *bytes.Buffer
	handlers writeHandlers
}

type JSONWriter struct {
	transmitWriter
}

func NewJSONWriter(buffer *bytes.Buffer) JSONWriter {
	handlers := map[reflect.Type]WriteHandler{
		reflect.TypeOf(uuid.NewV4()): uuidHandler(),
		reflect.TypeOf(2):            integerHandler(),
		reflect.TypeOf(&url.URL{}):   urlHandler(),
	}
	return JSONWriter{transmitWriter{buffer: buffer, handlers: handlers}}
}

func (w transmitWriter) lookupHandler(obj interface{}) (WriteHandler, error) {
	objType := reflect.TypeOf(obj)
	result, ok := w.handlers[objType]
	if !ok {
		return WriteHandler{}, fmt.Errorf("No handler found for type %+v", objType)
	}
	return result, nil
}

func (w JSONWriter) Write(obj interface{}) error {
	handler, err := w.lookupHandler(obj)

	if err != nil {
		return fmt.Errorf("Could not write obj without a handler for its type: %+v with type %s", obj, reflect.TypeOf(obj))
	}

	representation := handler.Rep(obj)

	switch t := representation.(type) {
	case string:
		_, err = w.buffer.WriteString(fmt.Sprintf("~%s%s", handler.Tag, representation))
	case int:
		_, err = w.buffer.WriteString(fmt.Sprintf("%d", representation))
	default:
		err = fmt.Errorf("Do not know what to do with the result of the handler for type %v", t)
	}

	return err
}

func uuidHandler() WriteHandler {
	return WriteHandler{
		Tag: "u",
		Rep: func(obj interface{}) interface{} {
			uuid, _ := obj.(uuid.UUID)
			return fmt.Sprintf("%s", uuid.String())
		},
	}
}

func integerHandler() WriteHandler {
	return WriteHandler{
		Tag: "i",
		Rep: func(obj interface{}) interface{} {
			integer, _ := obj.(int)
			if integer < 2*53 {
				return integer
			} else {
				return fmt.Sprintf("%d", integer)
			}
		},
	}
}

func urlHandler() WriteHandler {
	return WriteHandler{
		Tag: "r",
		Rep: func(obj interface{}) interface{} {
			url, _ := obj.(*url.URL)
			return url.String()
		},
	}
}
