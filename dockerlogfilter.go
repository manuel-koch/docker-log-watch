package main

import (
	"io"
	"unicode/utf8"
)

type DockerLogFilter struct {
	stream io.Reader
	data   []byte
	skip   int // skip number of read bytes from stream
}

// DockerLineHeaderSize denotes the size of the header for every line of docker log output.
// We want to skip fixed number of "header" bytes.
const DockerLineHeaderSize = 8

func NewDockerLogFilter(inputStream io.Reader) *DockerLogFilter {
	return &DockerLogFilter{inputStream, make([]byte, 0),
		DockerLineHeaderSize, // skip the first 8 bytes
	}
}

func (filter *DockerLogFilter) readUpstream(n int) {
	buf := make([]byte, n)
	readBytes, _ := filter.stream.Read(buf)
	if readBytes > 0 {
		filter.data = append(filter.data, buf[:readBytes]...)
	}

	if filter.skip > 0 {
		filter.skipData(filter.skip)
	}

	for i, w := 0, 0; i < len(filter.data); i += w {
		runeValue, width := utf8.DecodeRune(filter.data[i:])
		w = width
		if runeValue == 10 { // skip the next 8 bytes after a linefeed
			filter.skipDataAtOffset(DockerLineHeaderSize, i+1)
			w += DockerLineHeaderSize
		}
	}
}

func (filter *DockerLogFilter) skipDataAtOffset(n int, offset int) {
	filter.skip = n
	oldLen := len(filter.data)
	if offset+filter.skip < oldLen {
		copy(filter.data[offset:], filter.data[offset+filter.skip:])
		filter.data = filter.data[:oldLen-filter.skip]
	} else {
		filter.data = filter.data[:offset]
	}
	newLen := len(filter.data)
	filter.skip -= oldLen - newLen
}

func (filter *DockerLogFilter) skipData(n int) {
	filter.skipDataAtOffset(n, 0)
}

func (filter *DockerLogFilter) eof() bool {
	return len(filter.data) == 0
}

func (filter *DockerLogFilter) readByte() byte {
	// eof() check must already be done
	b := filter.data[0]
	filter.data = filter.data[1:]
	return b
}

func (filter *DockerLogFilter) Read(p []byte) (n int, err error) {
	filter.readUpstream(len(p))

	if filter.eof() {
		err = io.EOF
		return
	}

	if l := len(p); l > 0 {
		for n < l {
			p[n] = filter.readByte()
			n++
			if filter.eof() {
				break
			}
		}
	}
	return
}
