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
		result := readString("[\"^ \",\"key\",12]").(map[*MapKey]interface{})
		expected := map[string]int{"key": 12}

		for mapKey, v := range result {
			actualKey := mapKey.Key.(string)
			Expect(v).To(Equal(expected[actualKey]))
		}
	})

	It("reads a non-stringable simple map", func() {
		r := readString("[\"^ \",\"~i1\",\"hello\", \"~i2\", \"world\"]")
		result := r.(map[*MapKey]interface{})
		expected := map[int]string{1: "hello", 2: "world"}

		for mapKey, v := range result {
			actualKey := mapKey.Key.(int)
			Expect(v).To(Equal(expected[actualKey]))
		}
	})

	It("reads a simple map with cached keys", func() {
		result := readString("[[\"^ \",\"name\",\"JW\",\"town\",\"Enschede\"],[\"^ \",\"^0\",\"JW\",\"^1\",\"Enschede\"],[\"^ \",\"^0\",\"JW\",\"^1\",\"Enschede\"]]")

		m := map[interface{}]interface{}{"name": "JW", "town": "Enschede"}
		resultSlice, ok := result.([]interface{})
		Expect(ok)
		Expect(len(resultSlice)).To(Equal(3))
		for _, v := range resultSlice {
			subResult := v.(map[*MapKey]interface{})

			for mapKey, val := range subResult {
				realKey := mapKey.Key.(string)
				Expect(val).To(Equal(m[realKey]))
			}
		}
	})

	It("reads a complex map", func() {
		result := readString("[\"~#cmap\",[[1,2,3],\"~bZ29vZGJ5ZQ==\",[7,8,9],\"~bY3J1ZWw=\",[13,14,15],\"~bd29ybGQ=\"]]")

		resultMap, ok := result.(map[*MapKey]interface{})
		Expect(ok)
		Expect(len(resultMap)).To(Equal(3))

		for k, v := range resultMap {
			realKey := k.Key.([]interface{})
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

	It("turns unknown types into tagged values", func() {
		result := readString("[\"~#point\",[\"^ \",\"x\",3.140000104904175,\"y\",100.0]]")
		tv, ok := result.(TaggedValue)
		Expect(ok)

		Expect(tv.Tag).To(Equal("point"))
		valMap, ok := tv.Rep.(map[*MapKey]interface{})
		Expect(ok)
		Expect(len(valMap)).To(Equal(2))

		expected := map[string]float64{
			"x": 3.140000104904175,
			"y": 100.0,
		}

		for k, v := range valMap {
			realKey := k.Key.(string)
			Expect(v).To(Equal(expected[realKey]))
		}
	})

	It("allows custom readers", func() {
		type Point struct {
			x float32
			y float32
		}

		pointReader := ReadHandler{
			FromRep: func(rep interface{}) (interface{}, error) {
				repAsMap, ok := rep.(map[*MapKey]interface{})
				if !ok {
					return nil, fmt.Errorf("Expected to be able to type assert to map[*MapKey]interface{}")
				}
				pointMap := make(map[string]interface{})
				for mapKey, v := range repAsMap {
					pointMap[mapKey.Key.(string)] = v
				}

				res := Point{x: float32(pointMap["x"].(float64)), y: float32(pointMap["y"].(float64))}

				return res, nil
			},
		}

		point := Point{x: 3.14, y: 100}

		buffer := bytes.NewBufferString("[\"~#point\",[\"^ \",\"x\",3.140000104904175,\"y\",100.0]]")
		customHandlers := ReadHandlerMap{
			"point": pointReader,
		}
		reader := NewJSONReaderWithHandlers(buffer, customHandlers)
		result := reader.Read()

		resultAsPoint, ok := result.(Point)
		Expect(ok)
		Expect(resultAsPoint).To(Equal(point))
	})
})
