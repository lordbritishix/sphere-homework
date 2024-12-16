package integration_tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSphereSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sphere Test Suite")
}
