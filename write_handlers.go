package transit_go

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/twinj/uuid"
)

func arrayWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Array Write Handler",
		Tag:  func(obj interface{}) string { return "array" },
		Rep: func(obj interface{}) interface{} {
			return obj
		},
	}
}

func binaryWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Binary Write Handler",
		Tag:  func(obj interface{}) string { return "b" },
		Rep: func(obj interface{}) interface{} {
			return obj
		},
	}
}

func booleanWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Boolean Write Handler",
		Tag:  func(obj interface{}) string { return "?" },
		Rep: func(obj interface{}) interface{} {
			return obj
		},
		StringRep: func(obj interface{}) *string {
			str := strconv.FormatBool(obj.(bool))
			return &str
		},
	}
}

func keywordWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Keyword Write Handler",
		Tag:  func(obj interface{}) string { return ":" },
		Rep: func(obj interface{}) interface{} {
			strRep, _ := obj.(string)
			return strRep[1:len(strRep)]
		},
		StringRep: func(obj interface{}) *string {
			strRep, _ := obj.(string)
			str := strRep[1:len(strRep)]
			return &str
		},
	}
}

func stringableKeys(keys []reflect.Value, tagProvider TagProvider) bool {
	for _, key := range keys {
		actualKey := key.Interface()
		tag := tagProvider.GetTag(actualKey)

		_, keyIsAString := actualKey.(string)

		if len(tag) > 1 {
			return false
		} else if tag == "" && !keyIsAString {
			return false
		}
	}
	return true
}

func mapWriteHandler(tagProvider TagProvider) WriteHandler {
	return WriteHandler{
		Tag: func(obj interface{}) string {
			mapAsValue := reflect.ValueOf(obj)
			keys := mapAsValue.MapKeys()

			if stringableKeys(keys, tagProvider) {
				return "map"
			} else {
				return "cmap"
			}
		},
		Rep: func(obj interface{}) interface{} {
			mapAsValue := reflect.ValueOf(obj)
			keys := mapAsValue.MapKeys()

			if stringableKeys(keys, tagProvider) {
				entries, _ := mapToMapEntries(obj)
				return entries
			} else {
				var list []interface{}
				for _, key := range keys {
					list = append(list, key.Interface())
					list = append(list, mapAsValue.MapIndex(key).Interface())
				}
				tv := TaggedValue{Tag: "array", Rep: list}
				return tv
			}
		},
	}
}

func setWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Set Write Handler",
		Tag:  func(obj interface{}) string { return "set" },
		Rep: func(obj interface{}) interface{} {
			return TaggedValue{Tag: "array", Rep: obj}
		},
	}
}

func nilWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Nil Write Handler",
		Tag:  func(obj interface{}) string { return "_" },
		Rep: func(obj interface{}) interface{} {
			return nil
		},
		StringRep: func(obj interface{}) *string {
			str := ""
			return &str
		},
	}
}

func interfaceToFloat64(obj interface{}) float64 {
	var fl float64

	if fl32, ok := obj.(float32); ok {
		fl = float64(fl32)
	} else {
		fl = obj.(float64)
	}

	return fl
}

func floatWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Float Write Handler",
		Tag: func(obj interface{}) string {
			fl := interfaceToFloat64(obj)
			if math.IsNaN(fl) || math.IsInf(fl, 0) {
				return "z"
			} else {
				return "d"
			}
		},
		Rep: func(obj interface{}) interface{} {
			fl := interfaceToFloat64(obj)
			if math.IsNaN(fl) {
				return "NaN"
			} else if math.IsInf(fl, 1) {
				return "INF"
			} else if math.IsInf(fl, -1) {
				return "-INF"
			} else {
				return fl
			}
		},
		StringRep: func(obj interface{}) *string {
			str := strconv.FormatFloat(obj.(float64), 'f', -1, 64)
			return &str
		},
	}
}

func doubleWriteHandler() WriteHandler {
	return floatWriteHandler()
}

func integerWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Integer Write Handler",
		Tag:  func(obj interface{}) string { return "i" },
		Rep: func(obj interface{}) interface{} {
			return obj
		},
		StringRep: func(obj interface{}) *string {
			str := fmt.Sprintf("%d", obj)
			return &str
		},
	}
}

func quoteWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Quote Write Handler",
		Tag:  func(obj interface{}) string { return "'" },
		Rep: func(obj interface{}) interface{} {
			q := obj.(Quote)
			return q.Object
		},
		StringRep: func(obj interface{}) *string {
			panic("Cannot string rep quoted object")
		},
	}
}

func ratioWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Rational Write Handler",
		Tag:  func(obj interface{}) string { return "ratio" },
		Rep: func(obj interface{}) interface{} {
			rational := obj.(big.Rat)
			var list []int64
			list = append(list, rational.Num().Int64())
			list = append(list, rational.Denom().Int64())
			return list
		},
	}
}

func taggedValueWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "TaggedValue Write Handler",
		Tag: func(obj interface{}) string {
			tv := obj.(TaggedValue)
			return tv.Tag
		},
		Rep: func(obj interface{}) interface{} {
			tv := obj.(TaggedValue)
			return tv.Rep
		},
		StringRep: func(obj interface{}) *string {
			return nil
		},
	}
}

func timeWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Time Write Handler",
		Tag:  func(obj interface{}) string { return "m" },
		Rep: func(obj interface{}) interface{} {
			t := obj.(time.Time)
			utcTime := t.UTC()
			utcTimeNano := utcTime.UnixNano()
			return utcTimeNano / int64(time.Millisecond)
		},
		StringRep: func(obj interface{}) *string {
			t := obj.(time.Time)
			utcTime := t.UTC()
			utcTimeNano := utcTime.UnixNano()
			str := fmt.Sprintf("%d", utcTimeNano/int64(time.Millisecond))
			return &str
		},
		VerboseHandler: &WriteHandler{
			Tag: func(obj interface{}) string { return "t" },
			Rep: func(obj interface{}) interface{} {
				t := obj.(time.Time)
				utcTime := t.UTC()
				return fmt.Sprintf("%v", utcTime)
			},
		},
	}
}

func toStringWriteHandler(t string) WriteHandler {
	return WriteHandler{
		Name: "toString Write Handler",
		Tag:  func(obj interface{}) string { return t },
		Rep: func(obj interface{}) interface{} {
			return fmt.Sprintf("%v", obj)
		},
		StringRep: func(obj interface{}) *string {
			str := fmt.Sprintf("%v", obj)
			return &str
		},
	}
}

func runeWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Rune write handler",
		Tag:  func(obj interface{}) string { return "c" },
		Rep: func(obj interface{}) interface{} {
			r := obj.(rune)
			return string(r)
		},
	}
}

func read_int64(data []byte) (ret int64) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func uuidWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "UUID Write Handler",
		Tag:  func(obj interface{}) string { return "u" },
		Rep: func(obj interface{}) interface{} {
			uuid, _ := obj.(uuid.UUID)
			bytes := uuid.Bytes()

			var list []int64
			list = append(list, read_int64(bytes[0:7]))
			list = append(list, read_int64(bytes[8:15]))
			return list
		},
		StringRep: func(obj interface{}) *string {
			uuid, _ := obj.(uuid.UUID)
			str := uuid.String()
			return &str
		},
	}
}

func linkWriteHandler() WriteHandler {
	return WriteHandler{
		Name: "Link Write Handler",
		Tag:  func(obj interface{}) string { return "link" },
		Rep: func(obj interface{}) interface{} {
			linkMap, _ := obj.(map[interface{}]interface{})
			return linkMap
		},
	}
}
