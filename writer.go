package transit_go

import (
	"bytes"
	"fmt"
	"math/big"
	"net/url"
	"reflect"
	"time"

	"github.com/twinj/uuid"
)

type TransmitWriter interface {
	Write(interface{}) error
	Buffer() *bytes.Buffer
}

type TagProvider interface {
	GetTag(obj interface{}) string
}

type TagProviderAware interface {
	SetTagProvider(tp *TagProvider)
}

type WriteHandler struct {
	Name           string
	Tag            func(interface{}) string
	Rep            func(interface{}) interface{}
	StringRep      func(interface{}) *string
	VerboseHandler *WriteHandler
}

type MapWriteHandler struct {
	WriteHandler
	tagProvider *TagProvider
}

func (m *MapWriteHandler) SetTagProvider(tp *TagProvider) {
	m.tagProvider = tp
}

type WriteHandlerMap map[reflect.Type]WriteHandler

type transmitWriter struct {
	emitter  Emitter
	buffer   *bytes.Buffer
	handlers WriteHandlerMap
}

type JSONWriter struct {
	transmitWriter
}

func (w transmitWriter) Buffer() *bytes.Buffer {
	return w.buffer
}

func defaultWriteHandlers() WriteHandlerMap {
	integerHandler := integerWriteHandler()
	uriHandler := toStringWriteHandler("r")
	handlers := WriteHandlerMap{
		reflect.TypeOf(nil):                   nilWriteHandler(),
		reflect.TypeOf(true):                  booleanWriteHandler(),
		reflect.TypeOf(""):                    toStringWriteHandler("s"),
		reflect.TypeOf(2):                     integerHandler,
		reflect.TypeOf(int64(2)):              integerHandler,
		reflect.TypeOf(int32(2)):              integerHandler,
		reflect.TypeOf(3.14159265359):         floatWriteHandler(),
		reflect.TypeOf(float32(3.141)):        floatWriteHandler(),
		reflect.TypeOf(big.NewInt(2)):         toStringWriteHandler("f"),
		reflect.TypeOf('c'):                   runeWriteHandler(),
		reflect.TypeOf([]byte{}):              binaryWriteHandler(),
		reflect.TypeOf(uuid.NewV4()):          uuidWriteHandler(),
		reflect.TypeOf(&url.URL{}):            uriHandler,
		reflect.TypeOf([]int{}):               arrayWriteHandler(),
		reflect.TypeOf([]int32{}):             arrayWriteHandler(),
		reflect.TypeOf([]int64{}):             arrayWriteHandler(),
		reflect.TypeOf([]string{}):            arrayWriteHandler(),
		reflect.TypeOf([]float32{}):           arrayWriteHandler(),
		reflect.TypeOf([]float64{}):           arrayWriteHandler(),
		reflect.TypeOf([]bool{}):              arrayWriteHandler(),
		reflect.TypeOf([]rune{}):              arrayWriteHandler(),
		reflect.TypeOf([]map[string]string{}): arrayWriteHandler(),
		reflect.TypeOf([]interface{}{}):       arrayWriteHandler(),
		reflect.TypeOf(NewSet()):              setWriteHandler(),
		reflect.TypeOf(time.Now()):            timeWriteHandler(),
		reflect.TypeOf(big.Rat{}):             ratioWriteHandler(),
		reflect.TypeOf(Quote{}):               quoteWriteHandler(),
		reflect.TypeOf(TaggedValue{}):         taggedValueWriteHandler(),
	}
	return handlers
}

func (m WriteHandlerMap) GetTag(obj interface{}) string {
	handler, err := m.lookupHandler(obj)
	if err != nil {
		return ""
	}
	return handler.Tag(obj)
}

func NewJSONWriter(buffer *bytes.Buffer) JSONWriter {
	emitter := NewJsonEmitter(buffer, defaultWriteHandlers())

	return JSONWriter{transmitWriter{buffer: buffer, emitter: emitter, handlers: defaultWriteHandlers()}}
}

func NewJSONWriterWithHandlers(buffer *bytes.Buffer, customHandlers map[reflect.Type]WriteHandler) JSONWriter {
	handlers := defaultWriteHandlers()

	for typ, handler := range customHandlers {
		handlers[typ] = handler
	}

	emitter := NewJsonEmitter(buffer, handlers)
	return JSONWriter{transmitWriter{buffer: buffer, emitter: emitter, handlers: handlers}}
}

func (m WriteHandlerMap) lookupHandler(obj interface{}) (WriteHandler, error) {
	objType := reflect.TypeOf(obj)
	result, ok := m[objType]
	if !ok {

		// Maybe it is some kind of map
		if objType.Kind() == reflect.Map {
			return mapWriteHandler(m), nil
		} else if objType.Kind() == reflect.Array {
			return arrayWriteHandler(), nil
		}

		return WriteHandler{}, fmt.Errorf("No handler found for type %+v", objType)
	}
	return result, nil
}

func (w JSONWriter) Write(obj interface{}) error {
	return w.emitter.emit(obj, false, NewWriteCache(true))
}
