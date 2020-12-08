package keyconjurer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpApi(t *testing.T) {
	ProdAPI = "https://prod.keyconjurer.local"
	DevAPI = "https://dev.keyconjurer.local"

	Dev = false
	assert.Equal(t, "https://prod.keyconjurer.local/test", createAPIURL("/test"), "prod url should be properly formatted")
	assert.Equal(t, "https://prod.keyconjurer.local/test", createAPIURL("test"), "prod url should be properly formatted ")

	Dev = true
	assert.Equal(t, "https://dev.keyconjurer.local/test", createAPIURL("test"), "dev url should be properly formatted ")
	assert.Equal(t, "https://dev.keyconjurer.local/test", createAPIURL("test"), "dev url should be properly formatted ")
}
