package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderLineParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	v, _ := headers.Get("Host")
	assert.Equal(t, "localhost:42069", v)
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid multiple field-values
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:42069\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	v, _ = headers.Get("Host")
	assert.Equal(t, "localhost:42069,localhost:42069", v)
	assert.True(t, done)
}