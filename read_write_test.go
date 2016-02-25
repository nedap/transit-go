package transit_go

import (
	"bytes"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/twinj/uuid"
)

var _ = Describe("Write & Read", func() {
	var write = func(obj interface{}) string {
		var buffer bytes.Buffer
		writer := NewJSONWriter(&buffer)
		err := writer.Write(obj)
		Expect(err).To(BeNil())
		result := string(writer.Buffer().Bytes())
		return result
	}

	var read = func(str string) interface{} {
		buffer := bytes.NewBufferString(str)
		reader := NewJSONReader(buffer)
		result := reader.Read()
		return result
	}

	It("Can use the same buffer", func() {
		obj := "hi there"
		var buffer bytes.Buffer
		writer := NewJSONWriter(&buffer)
		err := writer.Write(obj)
		Expect(err).To(BeNil())
		reader := NewJSONReader(&buffer)
		result := reader.Read()

		Expect(result).To(Equal(obj))
	})

	It("roundtrips nil values", func() {
		var val interface{}
		val = nil
		result := read(write(val))
		Expect(result).To(BeNil())
	})

	It("roundtrips string values", func() {
		val := "hello"
		result := read(write(val))
		Expect(result).To(Equal(val))
	})

	It("roundtrips small integers", func() {
		val := 24
		result := read(write(val))
		Expect(result).To(Equal(val))
	})

	It("roundtrips big integers", func() {
		val := 9007199254740999
		result := read(write(val))
		Expect(result).To(Equal(val))
	})

	It("roundtrips floats", func() {
		val := 3.14159265359
		result := read(write(val))
		Expect(result).To(Equal(val))
	})

	It("roundtrips byte slices", func() {
		val := []byte("hello world")
		result := read(write(val))
		Expect(result).To(Equal(val))
	})

	It("roundtrips times", func() {
		timeInMillis := 1456231033010
		t := time.Unix(0, int64(timeInMillis)*int64(time.Millisecond))
		result := read(write(t))
		Expect(result).To(Equal(t))
	})

	It("roundtrips runes", func() {
		val := 'a'
		result := read(write(val))
		Expect(result).To(Equal(val))
	})

	It("roundtrips a UUID", func() {
		val, err := uuid.Parse("dda5a83f-8f9d-4194-ae88-5745c8ca94a7")
		Expect(err).To(BeNil())

		result := read(write(val))
		Expect(result).To(Equal(val))
	})

	It("roundtrips urls", func() {
		val, err := url.Parse("http://example.com/search")
		Expect(err).To(BeNil())
		result := read(write(val))
		Expect(result).To(Equal(val))
	})

	It("roundtrips simple int array", func() {
		val := []int{1, 2, 3, 4}
		result := (read(write(val))).([]interface{})

		for i, v := range result {
			Expect(v).To(Equal(val[i]))
		}
	})

	It("roundtrips simple map", func() {
		val := map[string]int{"key": 12}
		result := (read(write(val))).(map[interface{}]interface{})

		for k, v := range val {
			resV, found := result[k]
			Expect(found)
			Expect(v).To(Equal(resV))
		}
	})

	It("roundtrips a non-stringable key map", func() {
		val := map[int]string{1: "hello", 2: "world"}
		result := (read(write(val))).(map[interface{}]interface{})

		for k, v := range val {
			resV, found := result[k]
			Expect(found)
			Expect(v).To(Equal(resV))
		}
	})

	It("roundtrips a map with an intslice as key and byte slices as values", func() {
		val := map[[3]int][]byte{
			[3]int{1, 2, 3}:    []byte("goodbye"),
			[3]int{7, 8, 9}:    []byte("cruel"),
			[3]int{13, 14, 15}: []byte("world"),
		}

		result := (read(write(val))).(map[*MapKey]interface{})
		for k, v := range result {
			actualKey := (*k).key.([]interface{})
			Expect(len(actualKey)).To(Equal(3))
			testArr := [3]int{actualKey[0].(int), actualKey[1].(int), actualKey[2].(int)}
			Expect(val[testArr]).To(Equal(v))
		}
	})

})
