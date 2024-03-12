//go:build !solution

package otp

import (
	"io"
)

type MyReader struct {
	reader     io.Reader
	randomizer io.Reader
}

func (m *MyReader) Read(p []byte) (n int, e error) {
	blockSize := min(len(p), 1024)
	curBlock := make([]byte, blockSize)
	rndBlock := make([]byte, blockSize)
	e = nil
	n = 0
	curP := p
	curN := -1
	for curN != 0 && e == nil && n != len(p) {
		toXor := min(len(curP), blockSize)
		curN, e = m.reader.Read(curBlock[:toXor])
		toXor = min(toXor, curN)
		_, err := m.randomizer.Read(rndBlock[:toXor])
		if err != nil {
			return 0, err
		}
		for i := 0; i < toXor; i++ {
			curP[i] = curBlock[i] ^ rndBlock[i]
		}
		curP = curP[toXor:]
		n += toXor
	}
	return
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &MyReader{r, prng}
}

type MyWriter struct {
	writer     io.Writer
	randomizer io.Reader
}

func (m *MyWriter) Write(p []byte) (n int, e error) {
	const blockSize int = 1024
	curBlock := make([]byte, blockSize)
	rndBlock := make([]byte, blockSize)
	e = nil
	n = 0
	curP := p
	curN := -1
	for !(curN == 0 || e != nil || n == len(p)) {
		toXor := min(len(curP), blockSize)
		curN, _ = m.randomizer.Read(rndBlock)
		toXor = min(toXor, curN)
		for i := 0; i < toXor; i++ {
			curBlock[i] = curP[i] ^ rndBlock[i]
		}
		curN, e = m.writer.Write(curBlock[:toXor])
		toXor = min(toXor, curN)
		curP = curP[toXor:]
		n += toXor
	}
	return
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &MyWriter{w, prng}
}
