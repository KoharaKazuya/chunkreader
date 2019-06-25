package chunkreader_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/KoharaKazuya/chunkreader"
)

type chunkerMock struct {
	callCount int
	inner     io.Reader
}

func (c *chunkerMock) ReadChunk() ([]byte, error) {
	c.callCount++

	buf := make([]byte, 8)
	if c.inner != nil {
		n, err := c.inner.Read(buf)
		return buf[:n], err
	}
	return buf, nil
}

func TestNewInstance(t *testing.T) {
	c := &chunkerMock{}
	r := chunkreader.New(c)
	_ = r
}

func TestCallChunker(t *testing.T) {
	c := &chunkerMock{}
	r := chunkreader.New(c)

	buf := make([]byte, 1024)
	_, _ = r.Read(buf)

	if c.callCount == 0 {
		t.Errorf("mockReader didn't called")
	}
}

func TestTransparency(t *testing.T) {
	d := []byte{0x00, 0x01, 0x02, 0x03}
	c := &chunkerMock{
		inner: bytes.NewBuffer(d),
	}
	r := chunkreader.New(c)

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		t.Errorf("failed to read: %v", err)
	}

	if !bytes.Equal(buf[:n], d) {
		t.Errorf("unexpected read (n: %v, buf: %v)", n, buf)
	}

	_, err = r.Read(buf)
	if err != io.EOF {
		t.Errorf("cannot find EOF: %v", err)
	}
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestErrorTransparency(t *testing.T) {
	c := &chunkerMock{inner: &errorReader{}}
	r := chunkreader.New(c)

	buf := make([]byte, 1024)
	_, err := r.Read(buf)
	if err != io.ErrClosedPipe {
		t.Errorf("cannot find ErrClosedPipe: %v", err)
	}
}

type converter struct {
	inner     io.Reader
	chunkSize int
}

func (c *converter) ReadChunk() ([]byte, error) {
	buf := make([]byte, c.chunkSize)
	n, _ := io.ReadFull(c.inner, buf)
	if n == 0 {
		return nil, io.EOF
	}

	buf = buf[:n]

	for i := range buf {
		buf[i]++
	}

	return buf, nil
}

func TestConvert(t *testing.T) {
	data := []struct {
		inner  io.Reader
		size   int
		expect []byte
	}{
		{bytes.NewBuffer([]byte("abc")), 2, []byte("bcd")},
		{bytes.NewBuffer([]byte("abc")), 3, []byte("bcd")},
		{bytes.NewBuffer([]byte("abc")), 8, []byte("bcd")},
		{bytes.NewBuffer([]byte("ab")), 16, []byte("bc")},
		{bytes.NewBuffer([]byte("abcdefghijklmn")), 3, []byte("bcdefghi")},
		{bytes.NewBuffer([]byte("abcdefghijklmn")), 9, []byte("bcdefghi")},
	}

	for i, d := range data {
		r := chunkreader.New(&converter{d.inner, d.size})
		buf := make([]byte, 8)

		n, err := r.Read(buf)
		if err != nil {
			t.Errorf("failed to read: %v", err)
		}
		if !bytes.Equal(buf[:n], d.expect) {
			t.Errorf("[%d] unexpected bytes (expected: %v): %v", i, d.expect, buf[:n])
		}
	}
}

func TestRemain(t *testing.T) {
	inner := bytes.NewBuffer([]byte("abcdefgh"))
	r := chunkreader.New(&converter{inner, 8})
	buf := make([]byte, 8)

	r.Read(buf[:4])
	r.Read(buf[4:])

	if !bytes.Equal(buf, []byte("bcdefghi")) {
		t.Errorf("unexpected byte: %v", buf)
	}
}
