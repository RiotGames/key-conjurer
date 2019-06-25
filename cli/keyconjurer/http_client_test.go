package keyconjurer

import (
	"log"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	level, err := logrus.ParseLevel("debug")
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLevel(level)

	Logger = logger
}

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
