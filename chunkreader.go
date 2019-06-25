package chunkreader

import "io"

type Chunker interface {
	ReadChunk() ([]byte, error)
}

type chunkReader struct {
	inner  Chunker
	remain []byte
}

func New(chunker Chunker) io.Reader {
	return &chunkReader{inner: chunker}
}

func (r *chunkReader) Read(p []byte) (int, error) {
	n := copy(p, r.remain)
	r.remain = r.remain[n:]
	if n == len(p) {
		return n, nil
	}

	for {
		chunk, err := r.inner.ReadChunk()
		m := copy(p[n:], chunk)
		n += m

		if err != nil && err != io.EOF {
			return n, err
		}

		if n >= len(p) || err == io.EOF {
			r.remain = chunk[m:]
			break
		}
	}

	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}
