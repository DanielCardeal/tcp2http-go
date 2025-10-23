package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Single_Header_Success(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.False(t, done)
	assert.Equal(t, headers.Get("Host"), "localhost:42069")
	assert.Equal(t, headers.Get("HOST"), "localhost:42069")
	assert.Equal(t, 23, n)
}

func Test_Multiple_Headers_Success(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nUser-Agent: curl/8.14.1\r\n\r\n")

	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.False(t, done)
	assert.Equal(t, headers.Get("Host"), "localhost:42069")
	assert.Equal(t, 23, n)
	data = data[n:]

	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.False(t, done)
	assert.Equal(t, headers.Get("User-Agent"), "curl/8.14.1")
	assert.Equal(t, 25, n)
	data = data[n:]

	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.True(t, done)
	assert.Equal(t, 2, n)
}

func Test_Multiple_Value_Header_Success(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Set-Person: daniel-pessoa\r\nSet-Person: gabriel-moraes\r\n\r\n")

	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 27, n)
	assert.False(t, done)
	assert.Equal(t, headers.Get("Set-Person"), "daniel-pessoa")
	data = data[n:]

	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 28, n)
	assert.False(t, done)
	assert.Equal(t, headers.Get("Set-Person"), "daniel-pessoa,gabriel-moraes")
}

func Test_Trailing_Whitespace_Success(t *testing.T) {
	headers := NewHeaders()
	data := []byte("         Host: 		 localhost:42069				\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.False(t, done)
	assert.Equal(t, headers.Get("Host"), "localhost:42069")
	assert.Equal(t, 39, n)
}

func Test_Inner_Whitespace_Failure(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host : localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func Test_Empty_Field_Name_Failure(t *testing.T) {
	headers := NewHeaders()
	data := []byte(": localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func Test_Field_Name_Bad_Characters_Failure(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
