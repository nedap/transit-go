package transit_go

import (
	"bytes"
	"fmt"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/twinj/uuid"
)

var _ = Describe("JSON Reader", func() {
	var readString = func(str string) interface{} {
		buffer := bytes.NewBufferString(str)
		reader := NewJSONReader(buffer)
		result := reader.Read()
		return result
	}

	It("reads null", func() {
		result := readString("[\"~#'\",null]")
		Expect(result).To(BeNil())
	})

	It("reads a string", func() {
		result := readString("[\"~#'\", \"Hello world!\"]")
		Expect(result).To(Equal("Hello world!"))
	})

	It("reads a small int", func() {
		result := readString("[\"~#'\",24]")
		Expect(result).To(Equal(24))
	})

	It("reads a big int", func() {
		result := readString("[\"~#'\",\"~i9007199254740999\"]")
		Expect(result).To(Equal(9007199254740999))
	})

	It("reads a float", func() {
		result := readString("[\"~#'\",3.14159265359]")
		Expect(result).To(Equal(3.14159265359))
	})

	It("reads a simple byte slice", func() {
		result := readString("[\"~#'\",\"~baGVsbG8gd29ybGQ=\"]")
		Expect(result).To(Equal([]byte("hello world")))
	})

	It("reads a rune", func() {
		result := readString("[\"~#'\",\"~ca\"]")
		Expect(result).To(Equal('a'))
	})

	It("reads times", func() {
		timeInMillis := 1456231033010
		t := time.Unix(0, int64(timeInMillis)*int64(time.Millisecond))

		result := readString(fmt.Sprintf("[\"~#'\",\"~m%d\"]", timeInMillis))
		Expect(result).To(Equal(t))
	})

	It("reads a uuid", func() {
		result := readString("[\"~#'\",\"~udda5a83f-8f9d-4194-ae88-5745c8ca94a7\"]")
		uuid, err := uuid.Parse("dda5a83f-8f9d-4194-ae88-5745c8ca94a7")
		Expect(err).To(BeNil())

		Expect(result).To(Equal(uuid))
	})

	It("reads a url", func() {
		result := readString("[\"~#'\",\"~rhttp://example.com/search\"]")
		url, err := url.Parse("http://example.com/search")
		Expect(err).To(BeNil())

		Expect(result).To(Equal(url))
	})

	It("reads a simple int array", func() {
		result := readString("[1,2,3,4]")
		Expect(result).To(Equal([]interface{}{1, 2, 3, 4}))
	})

	It("reads a simple map", func() {
		result := readString("[\"^ \",\"key\",12]")
		Expect(result).To(Equal(map[interface{}]interface{}{"key": 12}))
	})

	It("reads a non-stringable simple map", func() {
		result := readString("[\"^ \",\"~i1\",\"hello\", \"~i2\", \"world\"]")
		Expect(result).To(Equal(map[interface{}]interface{}{1: "hello", 2: "world"}))
	})

	It("reads a simple map with cached keys", func() {
		result := readString("[[\"^ \",\"name\",\"JW\",\"town\",\"Enschede\"],[\"^ \",\"^0\",\"JW\",\"^1\",\"Enschede\"],[\"^ \",\"^0\",\"JW\",\"^1\",\"Enschede\"]]")

		m := map[interface{}]interface{}{"name": "JW", "town": "Enschede"}
		resultSlice, ok := result.([]interface{})
		Expect(ok)
		Expect(len(resultSlice)).To(Equal(3))
		for _, v := range resultSlice {
			subResult := v.(map[interface{}]interface{})
			Expect(subResult).To(Equal(m))
		}
	})

	It("reads a complex map", func() {
		result := readString("[\"~#cmap\",[[1,2,3],\"~bZ29vZGJ5ZQ==\",[7,8,9],\"~bY3J1ZWw=\",[13,14,15],\"~bd29ybGQ=\"]]")

		resultMap, ok := result.(map[*MapKey]interface{})
		Expect(ok)
		Expect(len(resultMap)).To(Equal(3))

		for k, v := range resultMap {
			realKey := k.Key().([]interface{})
			firstElem := realKey[0]
			firstInt, ok := firstElem.(int)
			Expect(ok)
			switch firstInt {
			case 1:
				Expect(v).To(Equal([]byte("goodbye")))
			case 7:
				Expect(v).To(Equal([]byte("cruel")))
			case 13:
				Expect(v).To(Equal([]byte("world")))
			}
		}
	})

	It("allows custom readers", func() {
		type Point struct {
			x float32
			y float32
		}

		pointReader := ReadHandler{
			FromRep: func(rep interface{}) (interface{}, error) {
				repAsMap, ok := rep.(map[interface{}]interface{})
				if !ok {
					return nil, fmt.Errorf("Expected to be able to type assert to map[interface{}]interface{}")
				}

				res := Point{x: float32(repAsMap["x"].(float64)), y: float32(repAsMap["y"].(float64))}

				return res, nil
			},
		}

		point := Point{x: 3.14, y: 100}

		buffer := bytes.NewBufferString("[\"~#point\",[\"^ \",\"x\",3.140000104904175,\"y\",100.0]]")
		customHandlers := map[string]ReadHandler{
			"point": pointReader,
		}
		reader := NewJSONReaderWithHandlers(buffer, customHandlers)
		result := reader.Read()

		resultAsPoint, ok := result.(Point)
		Expect(ok)
		Expect(resultAsPoint).To(Equal(point))
	})
})
