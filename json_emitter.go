package transit_go

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/nedap/transit-go/constants"
	"strconv"
)

const (
	jsonMaxInt = 2 ^ 53 - 1
	jsonMinInt = -jsonMaxInt
)

type JsonEmitter struct {
	buffer *bytes.Buffer
	base   baseEmitter
}

func NewJsonEmitter(buffer *bytes.Buffer, writeHandlerMap WriteHandlerMap) Emitter {
	jsonEmitter := &JsonEmitter{buffer: buffer}
	baseEmitter := baseEmitter{buffer: buffer, writeHandlerMap: writeHandlerMap, emitter: jsonEmitter}
	jsonEmitter.base = baseEmitter
	return jsonEmitter
}

func (j *JsonEmitter) emit(obj interface{}, asMapKey bool, cache WriteCache) error {
	return j.base.marshalTop(obj, cache)
}

func (j *JsonEmitter) emitNil(asMapKey bool, cache WriteCache) error {
	if asMapKey {
		err := j.emitString(constants.ESC_STR, "_", "", asMapKey, cache)
		if err != nil {
			return err
		}
	} else {
		j.buffer.WriteString("null")
		return nil
	}
	return nil
}

func maybePrefix(prefix, tag, str string) string {
	if prefix == "" && tag == "" {
		return str
	}
	return fmt.Sprintf("%s%s%s", prefix, tag, str)
}

func (j *JsonEmitter) emitString(prefix, tag, str string, asMapKey bool, cache WriteCache) error {
	outString := cache.CacheWrite(maybePrefix(prefix, tag, str), asMapKey)
	j.buffer.WriteString(fmt.Sprintf("\"%s\"", outString))
	return nil
}

func (j *JsonEmitter) emitBoolean(b bool, asMapKey bool, cache WriteCache) error {
	if asMapKey {
		var str string
		if b {
			str = "t"
		} else {
			str = "f"
		}
		return j.emitString(constants.ESC_STR, "?", str, asMapKey, cache)
	} else {
		j.buffer.WriteString(strconv.FormatBool(b))
		return nil
	}
}

func (j *JsonEmitter) emitInteger(intValue int, asMapKey bool, cache WriteCache) error {
	intStr := strconv.FormatInt(int64(intValue), 10)
	if asMapKey || intValue > jsonMaxInt || intValue < jsonMinInt {
		return j.emitString(constants.ESC_STR, "i", intStr, asMapKey, cache)
	} else {
		j.buffer.WriteString(intStr)
		return nil
	}
}

func (j *JsonEmitter) emitDouble(floatValue float64, asMapKey bool, cache WriteCache) error {
	floatStr := strconv.FormatFloat(floatValue, 'f', -1, 64)
	if asMapKey {
		return j.emitString(constants.ESC_STR, "d", floatStr, asMapKey, cache)
	} else {
		j.buffer.WriteString(floatStr)
		return nil
	}
}

func (j *JsonEmitter) emitBinary(bytes []byte, asMapKey bool, cache WriteCache) error {
	encodedBytes := base64.StdEncoding.EncodeToString(bytes)
	return j.emitString(constants.ESC_STR, "b", fmt.Sprintf("%s", encodedBytes), asMapKey, cache)
}

func (j *JsonEmitter) emitArrayStart(size int) error {
	j.buffer.WriteString("[")
	return nil
}

func (j *JsonEmitter) emitArrayEnd() error {
	j.buffer.WriteString("]")
	return nil
}

func (j *JsonEmitter) emitMapStart(size int) error {
	j.buffer.WriteString("{")
	return nil
}

func (j *JsonEmitter) emitMapEnd() error {
	j.buffer.WriteString("}")
	return nil
}

func (j *JsonEmitter) prefersStrings() bool {
	return true
}

func (j *JsonEmitter) flushWriter() error {
	return nil
}

func (j *JsonEmitter) emitActualMap(entries mapEntries, ignored bool, cache WriteCache) (err error) {
	size := entries.Len()
	err = j.emitArrayStart(size)
	if err != nil {
		return err
	}
	err = j.emitString("", "", constants.MAP_AS_ARRAY, false, cache)
	if size > 0 {
		j.buffer.WriteString(",")
	}
	if err != nil {
		return err
	}

	for index, entry := range entries.Items() {
		entry, ok := entry.(mapEntry)
		if ok {
			err = j.base.marshal(entry.key, true, cache)
			if err != nil {
				return err
			}
			j.buffer.WriteString(",")
			err = j.base.marshal(entry.value, false, cache)
			if err != nil {
				return err
			}

			if index < size-1 {
				j.buffer.WriteString(",")
			}
		}
	}
	err = j.emitArrayEnd()
	return err
}
