package transit_go

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/nedap/transit-go/constants"
)

type baseEmitter struct {
	buffer          *bytes.Buffer
	writeHandlerMap WriteHandlerMap
	emitter         Emitter
}

func escape(str string) string {
	length := len(str)
	if length > 0 {
		r := str[0]
		if r == constants.ESC || r == constants.SUB || r == constants.RESERVED {
			return fmt.Sprintf("%s%s", constants.ESC, str)
		}
	}
	return str
}

func (e *baseEmitter) emitTagged(t string, obj interface{}, ignored bool, cache WriteCache) error {
	err := e.emitter.emitArrayStart(2)
	if err != nil {
		return err
	}
	err = e.emitter.emitString(constants.ESC_TAG, t, "", false, cache)
	if err != nil {
		return err
	}
	e.buffer.WriteString(",")
	err = e.marshal(obj, false, cache)
	if err != nil {
		return err
	}
	err = e.emitter.emitArrayEnd()
	return err
}

func (e *baseEmitter) emitEncoded(t string, handler WriteHandler, obj interface{}, asMapKey bool, cache WriteCache) error {
	if len(t) == 1 {
		repr := handler.Rep(obj)

		// repr is an instance of a string
		if strRep, ok := repr.(string); ok {
			return e.emitter.emitString(constants.ESC_STR, t, strRep, asMapKey, cache)
		} else if e.emitter.prefersStrings() || asMapKey {
			sr := handler.StringRep(obj)
			if sr != nil {
				return e.emitter.emitString(constants.ESC_STR, t, *sr, asMapKey, cache)
			} else {
				return fmt.Errorf("%+v cannot be encoded as string", obj)
			}
		} else {
			return e.emitTagged(t, repr, asMapKey, cache)
		}
	} else if asMapKey {
		return fmt.Errorf("Cannot use %+v as a map key", obj)
	} else {
		return e.emitTagged(t, handler.Rep(obj), asMapKey, cache)
	}
}

func (e *baseEmitter) emitMap(m interface{}, ignored bool, cache WriteCache) error {

	entries, _ := m.(mapEntries)

	return e.emitter.emitActualMap(entries, ignored, cache)
}

func (e *baseEmitter) emitArray(obj interface{}, ignored bool, cache WriteCache) error {
	value := reflect.ValueOf(obj)
	kind := value.Kind()

	if kind != reflect.Slice && kind != reflect.Array {
		return fmt.Errorf("Cannot emit array when obj is not a slice; %+v", obj)
	}

	err := e.emitter.emitArrayStart(value.Len())
	if err != nil {
		return err
	}

	for i := 0; i < value.Len(); i++ {
		e.marshal(value.Index(i).Interface(), false, cache)
		if i < value.Len()-1 {
			e.buffer.WriteString(",")
		}
	}

	return e.emitter.emitArrayEnd()
}

func (e *baseEmitter) marshal(obj interface{}, asMapKey bool, cache WriteCache) error {
	var err error
	handler, err := e.writeHandlerMap.lookupHandler(obj)
	supported := false

	if err == nil {
		tag := handler.Tag(obj)

		if tag != "" {
			supported = true
			if len(tag) == 1 {
				switch tag[0] {
				case '_':
					err = e.emitter.emitNil(asMapKey, cache)
				case 's':
					err = e.emitter.emitString("", "", escape((handler.Rep(obj)).(string)), asMapKey, cache)
				case '?':
					err = e.emitter.emitBoolean((handler.Rep(obj)).(bool), asMapKey, cache)
				case 'i':
					err = e.emitter.emitInteger((handler.Rep(obj)).(int), asMapKey, cache)
				case 'd':
					err = e.emitter.emitDouble((handler.Rep(obj)).(float64), asMapKey, cache)
				case 'b':
					err = e.emitter.emitBinary((handler.Rep(obj)).([]byte), asMapKey, cache)
				case '\'':
					err = e.emitTagged(tag, handler.Rep(obj), false, cache)
				default:
					err = e.emitEncoded(tag, handler, obj, asMapKey, cache)
				}
			} else {
				switch tag {
				case "array":
					err = e.emitArray(handler.Rep(obj), asMapKey, cache)
				case "map":
					err = e.emitMap(handler.Rep(obj), asMapKey, cache)
				default:
					err = e.emitEncoded(tag, handler, obj, asMapKey, cache)
				}
			}
			if err != nil {
				return err
			}
			return e.emitter.flushWriter()
		}
		return err
	}
	if !supported {
		return fmt.Errorf("%s is not supported", reflect.TypeOf(obj).String())
	}
	return nil
}

func (e *baseEmitter) marshalTop(obj interface{}, cache WriteCache) error {
	object := obj
	handler, err := e.writeHandlerMap.lookupHandler(obj)
	if err != nil {
		return err
	}

	tag := handler.Tag(obj)
	if tag == "" {
		return fmt.Errorf("%s is not supported", reflect.TypeOf(obj).String())
	}

	if len(tag) == 1 {
		object = Quote{Object: obj}
	}

	return e.marshal(object, false, cache)
}
