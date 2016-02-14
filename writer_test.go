package transit_go

import (
	"bytes"
	"fmt"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/twinj/uuid"
)

var _ = Describe("Writer", func() {
	fmt.Println("At least we are described")
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	AfterEach(func() {
		buffer.Reset()
	})

	It("marshals a UUID", func() {
		uuid, err := uuid.Parse("dda5a83f-8f9d-4194-ae88-5745c8ca94a7")
		Expect(err).To(BeNil())

		err = writer.Write(uuid)
		Expect(err).To(BeNil())

		result := string(buffer.Bytes())
		Expect(result).To(Equal("~udda5a83f-8f9d-4194-ae88-5745c8ca94a7"))
	})

	It("marshals a big integer", func() {
		number := 9007199254740999
		err := writer.Write(number)
		Expect(err).To(BeNil())

		result := string(buffer.Bytes())
		Expect(result).To(Equal("~i9007199254740999"))
	})

	It("marshals small integers", func() {
		number := 24
		err := writer.Write(number)
		Expect(err).To(BeNil())

		result := string(buffer.Bytes())
		Expect(result).To(Equal("24"))
	})

	It("marshals an net/URL", func() {
		url, err := url.Parse("http://example.com/search")
		Expect(err).To(BeNil())

		err = writer.Write(url)
		Expect(err).To(BeNil())

		result := string(buffer.Bytes())
		Expect(result).To(Equal("~rhttp://example.com/search"))
	})
})
