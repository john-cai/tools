package mcredis_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRedisfoil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mcredis Suite")
}
