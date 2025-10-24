package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func Test_Standard_Get_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	assert.Equal(t, r.Headers.Get("Host"), "localhost:42069")
	assert.Equal(t, r.Headers.Get("User-Agent"), "curl/7.81.0")
	assert.Equal(t, r.Headers.Get("Accept"), "*/*")
}

func Test_Non_Empty_Target_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func Test_Standard_Post_Success(t *testing.T) {
	const BODY = `{"key":"value","pair":true}` + "\n"
	r, err := RequestFromReader(&chunkReader{
		data: "POST / HTTP/1.1\r\n" +
			"Content-Length: 28\r\n" +
			"\r\n" +
			BODY,
		numBytesPerRead: 5,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "28", r.Headers.Get("Content-Length"))
	assert.Equal(t, BODY, string(r.Body))
}

func Test_Zeroed_Content_Lenght_Empty_Body_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data: "POST /coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 5,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "0", r.Headers.Get("Content-Length"))
	assert.Empty(t, r.Body)
}

func Test_No_Content_Lenght_Empty_Body_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data: "POST /coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"\r\n",
		numBytesPerRead: 5,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Body)
}

func Test_Long_Body_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data: "POST / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Content-Length: 16\r\n" +
			"\r\n" +
			"x's will be gonexxxxxxxxxxxxx",
		numBytesPerRead: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "16", r.Headers.Get("Content-Length"))
	assert.Equal(t, "x's will be gone", string(r.Body))
}

func Test_Missing_Content_Length_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data: "POST / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"\r\n" +
			"this will be ignored (aka not parsed)",
		numBytesPerRead: 20,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Headers.Get("Content-Length"))
	assert.Empty(t, r.Body)
}

func Test_Case_Insensitive_Headers_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data:            "GET / HTTP/1.1\r\nHOST: localhost:42069\r\nuser-agent: curl/7.81.0\r\n\r\n",
		numBytesPerRead: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, r.Headers.Get("host"), "localhost:42069")
	assert.Equal(t, r.Headers.Get("USER-AGENT"), "curl/7.81.0")
}

func Test_Empty_Headers_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
}

func Test_Duplicate_Headers_Success(t *testing.T) {
	r, err := RequestFromReader(&chunkReader{
		data:            "GET / HTTP/1.1\r\nAdd-User: daniel\r\nAdd-User: miguel\r\n\r\n",
		numBytesPerRead: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, r.Headers.Get("Add-User"), "daniel,miguel")
}

func Test_Unsuported_HTTP_Version_Failure(t *testing.T) {
	_, err := RequestFromReader(&chunkReader{
		data:            "GET /coffee HTTP/2.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	})
	require.Error(t, err)
}

func Test_Missing_End_Of_Headers_Failure(t *testing.T) {
	_, err := RequestFromReader(&chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\n",
		numBytesPerRead: 3,
	})
	require.Error(t, err)
}

func Test_Missing_Request_Line_Parts_Failure(t *testing.T) {
	_, err := RequestFromReader(&chunkReader{
		data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	})
	require.Error(t, err)
}

func Test_Malformed_Header_Failure(t *testing.T) {
	_, err := RequestFromReader(&chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 4,
	})
	require.Error(t, err)
}

func Test_Short_Body_Failure(t *testing.T) {
	_, err := RequestFromReader(&chunkReader{
		data: "POST / HTTP/1.1\r\n" +
			"Content-Length: 512\r\n" +
			"\r\n" +
			"too short lol",
		numBytesPerRead: 12,
	})
	require.Error(t, err)
}
