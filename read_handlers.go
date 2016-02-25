package transit_go

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"strconv"
	"time"

	"github.com/twinj/uuid"
)

func bigDecimalReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Big Decimal",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			var bigFloat big.Float
			fl, _, err := bigFloat.Parse(strRep, 10)
			if err != nil {
				return nil, err
			}
			return fl, nil
		},
	}
}

func bigIntegerReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Big Integer",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			var bigInt *big.Int
			bigInt, ok := bigInt.SetString(strRep, 10)
			if !ok {
				return nil, fmt.Errorf("Could not convert '%s' to big.Int", strRep)
			}
			return bigInt, nil
		},
	}
}

func binaryReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Binary",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			bytes, err := base64.StdEncoding.DecodeString(strRep)
			if err != nil {
				return nil, err
			}
			return bytes, nil
		},
	}
}

func booleanReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Boolean",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			return strRep == "t", nil
		},
	}
}

func characterReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Character",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			return rune(strRep[0]), nil
		},
	}
}

/* cMapReadHandler */
type MapKey struct {
	key interface{}
}

func newMapKey(key interface{}) *MapKey {
	return &MapKey{key: key}
}

func (mk MapKey) Key() interface{} {
	k := mk.key
	return k
}

type cMapArrayReader struct {
	m       map[interface{}]interface{}
	nextKey interface{}
}

func (c *cMapArrayReader) Init(size int) interface{} {
	m := make(map[*MapKey]interface{})
	return m
}

func (c *cMapArrayReader) Add(a interface{}, item interface{}) interface{} {
	m, _ := a.(map[*MapKey]interface{})
	if c.nextKey != nil {
		mk := newMapKey(c.nextKey)
		m[mk] = item
		c.nextKey = nil
	} else {
		c.nextKey = item
	}
	return m
}

func (c *cMapArrayReader) Complete(a interface{}) interface{} {
	return a
}

func cmapReadHandler() ArrayReadHandler {
	rh := ReadHandler{
		Name: "cMap",
		FromRep: func(rep interface{}) (interface{}, error) {
			return nil, fmt.Errorf("'FromRep' is not supported")
		},
	}
	return ArrayReadHandler{ReadHandler: rh, arrayReader: new(cMapArrayReader)}
}

func doubleReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Double",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			return strconv.ParseFloat(strRep, 64)
		},
	}
}

func specialNumberReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Special Number",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			if strRep == "NaN" {
				return math.NaN(), nil
			} else if strRep == "INF" {
				return math.Inf(1), nil
			} else if strRep == "-INF" {
				return math.Inf(-1), nil
			} else {
				return nil, fmt.Errorf("Could not read special number")
			}
		},
	}
}

func identityReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Identity",
		FromRep: func(rep interface{}) (interface{}, error) {
			return rep, nil
		},
	}
}

func integerReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Integer",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			bigInt, err := strconv.ParseInt(strRep, 10, 64)
			if err != nil {
				return nil, err
			}
			return int(bigInt), nil
		},
	}
}

func keywordReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Keyword",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			return Keyword(strRep), nil
		},
	}
}

/* ListReadHandler */
type listArrayReader struct{}

func (c listArrayReader) Init(size int) interface{} {
	list := make([]interface{}, size)
	return list
}

func (c listArrayReader) Add(a interface{}, item interface{}) interface{} {
	list, _ := a.([]interface{})
	list = append(list, item)
	return list
}

func (c listArrayReader) Complete(a interface{}) interface{} {
	return a
}

func listReadHandler() ArrayReadHandler {
	rh := ReadHandler{
		Name: "List",
		FromRep: func(rep interface{}) (interface{}, error) {
			return nil, fmt.Errorf("'FromRep' is not supported")
		},
	}
	return ArrayReadHandler{ReadHandler: rh, arrayReader: listArrayReader{}}
}

func nullReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Null",
		FromRep: func(rep interface{}) (interface{}, error) {
			return nil, nil
		},
	}
}

func ratioReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Ratio",
		FromRep: func(rep interface{}) (interface{}, error) {
			list, _ := rep.([]int64)
			rational := big.NewRat(list[0], list[1])
			return rational, nil
		},
	}
}

/* SetReadHandler */
type setArrayReader struct{}

func (c setArrayReader) Init(size int) interface{} {
	return NewSet()
}

func (c setArrayReader) Add(a interface{}, item interface{}) interface{} {
	set, _ := a.(Set)
	set.Add(item)
	return set
}

func (c setArrayReader) Complete(a interface{}) interface{} {
	set, _ := a.(Set)
	return set
}

func setReadHandler() ArrayReadHandler {
	rh := ReadHandler{
		Name: "Set",
		FromRep: func(rep interface{}) (interface{}, error) {
			return nil, fmt.Errorf("'FromRep' is not supported")
		},
	}
	return ArrayReadHandler{ReadHandler: rh, arrayReader: setArrayReader{}}
}

func symbolReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Symbol",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			return Symbol(strRep), nil
		},
	}
}

func timeReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Time",
		FromRep: func(rep interface{}) (interface{}, error) {
			intRep, ok := rep.(int64)
			if !ok {
				strRep, _ := rep.(string)
				intRep, _ = strconv.ParseInt(strRep, 10, 64)
			}

			millis := intRep
			secs := millis / 1000
			rest := millis % 1000
			return time.Unix(secs, rest*int64(time.Millisecond)), nil
		},
	}
}

func uriReadHandler() ReadHandler {
	return ReadHandler{
		Name: "URI",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			return url.Parse(strRep)
		},
	}
}

func uuidReadHandler() ReadHandler {
	return ReadHandler{
		Name: "UUID",
		FromRep: func(rep interface{}) (interface{}, error) {
			strRep, _ := rep.(string)
			return uuid.Parse(strRep)
		},
	}
}

func linkReadHandler() ReadHandler {
	return ReadHandler{
		Name: "Link",
		FromRep: func(rep interface{}) (interface{}, error) {
			linkMap, ok := rep.(map[string]string)
			if !ok {
				return nil, fmt.Errorf("Could not convert representation to map[string]string")
			}
			return NewLinkFromMap(linkMap)
		},
	}
}
