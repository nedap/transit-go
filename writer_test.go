package transit_go

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/twinj/uuid"
)

func write(writer TransmitWriter, obj interface{}) string {
	err := writer.Write(obj)
	Expect(err).To(BeNil())
	result := string(writer.Buffer().Bytes())
	return result
}

var _ = Describe("JSON Writer", func() {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	AfterEach(func() {
		buffer.Reset()
	})

	It("marshals nil", func() {
		result := write(writer, nil)
		Expect(result).To(Equal("[\"~#'\",null]"))
	})

	It("marshals strings", func() {
		result := write(writer, "a string")
		Expect(result).To(Equal("[\"~#'\",\"a string\"]"))
	})

	It("marshals small integers", func() {
		number := 24
		result := write(writer, number)
		Expect(result).To(Equal("[\"~#'\",24]"))
	})

	It("marshals a big integer", func() {
		number := 9007199254740999

		result := write(writer, number)
		Expect(result).To(Equal("[\"~#'\",\"~i9007199254740999\"]"))
	})

	It("marshals a float", func() {
		pi := 3.14159265359
		result := write(writer, pi)
		Expect(result).To(Equal("[\"~#'\",3.14159265359]"))
	})

	It("marshals a single byte slice", func() {
		slice := []byte("hello world")
		result := write(writer, slice)
		Expect(result).To(Equal("[\"~#'\",\"~baGVsbG8gd29ybGQ=\"]"))
	})

	It("marshals a byte slice", func() {
		var random = func(n int) ([]byte, error) {
			b := make([]byte, n)
			_, err := rand.Read(b)
			if err != nil {
				return nil, err
			}
			return b, nil
		}

		randomBytes, err := random(128)
		Expect(err).To(BeNil())

		result := write(writer, randomBytes)

		encodedBytes := base64.StdEncoding.EncodeToString(randomBytes)

		Expect(result).To(Equal(fmt.Sprintf("[\"~#'\",\"%s\"]", fmt.Sprintf("~b%s", encodedBytes))))
	})

	It("marshals a time", func() {
		t := time.Now()

		result := write(writer, t)
		expected := t.UTC().UnixNano() / int64(time.Millisecond)

		Expect(result).To(Equal(fmt.Sprintf("[\"~#'\",\"%s\"]", fmt.Sprintf("~m%d", expected))))
	})

	It("marshals runes", func() {
		char := 'a'
		result := write(writer, char)
		Expect(result).To(Equal(fmt.Sprintf("[\"~#'\",\"%s\"]", fmt.Sprintf("~c%s", string(char)))))
	})

	It("marshals a UUID", func() {
		uuid, err := uuid.Parse("dda5a83f-8f9d-4194-ae88-5745c8ca94a7")
		Expect(err).To(BeNil())

		result := write(writer, uuid)
		Expect(result).To(Equal(fmt.Sprintf("[\"~#'\",\"%s\"]", "~udda5a83f-8f9d-4194-ae88-5745c8ca94a7")))
	})

	It("marshals an net/URL", func() {
		url, err := url.Parse("http://example.com/search")
		Expect(err).To(BeNil())

		result := write(writer, url)
		Expect(result).To(Equal(fmt.Sprintf("[\"~#'\",\"%s\"]", "~rhttp://example.com/search")))
	})

	It("marshals a simple int array", func() {
		arr := []int{1, 2, 3, 4}

		result := write(writer, arr)
		Expect(result).To(Equal("[1,2,3,4]"))
	})

	It("marshals a simple map", func() {
		m := map[string]int{"key": 12}

		result := write(writer, m)
		Expect(result).To(Equal("[\"^ \",\"key\",12]"))
	})

	It("marshals a non-stringable key map", func() {
		m := map[int]string{1: "hello", 2: "world"}

		result := write(writer, m)
		Expect(result).To(MatchRegexp("[\"^ \",(.+)]]"))
		Expect(result).To(MatchRegexp("\"~i1\",\"hello\""))
		Expect(result).To(MatchRegexp("\"~i2\",\"world\""))
	})

	It("marshals and caches stringable keys", func() {
		m := map[string]string{"name": "JW", "town": "Enschede"}
		a := []map[string]string{m, m, m}

		result := write(writer, a)

		Expect(result).To(MatchRegexp("\"\\^\\d\",\"JW\""))
		Expect(result).To(MatchRegexp("\"\\^\\d\",\"Enschede\""))
	})

	It("marshals nested maps", func() {
		m := map[string]interface{}{
			"id":     12,
			"action": "delete",
			"resource": map[string]interface{}{
				"owner_type":    "Store",
				"owner_id":      5,
				"resource_type": "Cleaner",
				"resource_id":   3,
			},
		}

		result := write(writer, m)
		// Test whether the result has a map as value for the 'resource' key
		Expect(result).To(MatchRegexp("\"resource\",\\[\"\\^ \","))
	})

	It("marshals a map with an intslice as key and byte slices as values", func() {
		m := map[[3]int][]byte{
			[3]int{1, 2, 3}:    []byte("goodbye"),
			[3]int{7, 8, 9}:    []byte("cruel"),
			[3]int{13, 14, 15}: []byte("world"),
		}

		result := write(writer, m)
		Expect(result).To(MatchRegexp("\\[\"~#cmap\",(.+)\\]\\]"))
		Expect(result).To(MatchRegexp("\\[1,2,3\\],\"~bZ29vZGJ5ZQ==\""))
		Expect(result).To(MatchRegexp("\\[7,8,9\\],\"~bY3J1ZWw=\""))
		Expect(result).To(MatchRegexp("\\[13,14,15\\],\"~bd29ybGQ=\""))
	})

	It("allows custom writers", func() {
		type Point struct {
			x float32
			y float32
		}

		pointHandler := WriteHandler{
			Name: "Point Write Handler",
			Tag:  func(obj interface{}) string { return "point" },
			Rep: func(obj interface{}) interface{} {
				p := obj.(Point)
				mapRepr := map[string]float32{"x": p.x, "y": p.y}
				return mapRepr
			},
		}

		customHandlers := map[reflect.Type]WriteHandler{
			reflect.TypeOf(Point{}): pointHandler,
		}

		var buffer bytes.Buffer
		writer := NewJSONWriterWithHandlers(&buffer, customHandlers)

		point := Point{x: 3.14, y: 100}

		err := writer.Write(point)
		Expect(err).To(BeNil())

		result := string(writer.Buffer().Bytes())

		Expect(result).To(MatchRegexp("\\[\"~#point\",(.+)\\]\\]"))
		Expect(result).To(MatchRegexp(",\\[\"\\^ \","))
		Expect(result).To(MatchRegexp("\"x\",3.140000104904175"))
		Expect(result).To(MatchRegexp("\"y\",100"))
	})
})
