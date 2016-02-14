package transit_go_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTransitGo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TransitGo Suite")
}
