package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAddress(t *testing.T) {
	addr, err := ParseAddress("https://idp.example.com")
	assert.Equal(t, "https://idp.example.com", addr.String())
	assert.NoError(t, err)
	addr, err = ParseAddress("https://idp.example.com:4000")
	assert.Equal(t, "https://idp.example.com:4000", addr.String())
	assert.NoError(t, err)
	addr, err = ParseAddress("http://idp.example.com:4000")
	assert.Equal(t, "http://idp.example.com:4000", addr.String())
	assert.NoError(t, err)
	addr, err = ParseAddress("localhost:4000")
	assert.Equal(t, "http://localhost:4000", addr.String())
	assert.NoError(t, err)
	addr, err = ParseAddress("localhost")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost", addr.String())
	addr, err = ParseAddress("127.0.0.1:4000")
	assert.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:4000", addr.String())
	_, err = ParseAddress("localhost:4000/foo")
	assert.ErrorIs(t, errHostnameCannotContainPath, err)
	_, err = ParseAddress("localhost/foo")
	assert.ErrorIs(t, errHostnameCannotContainPath, err)
}
