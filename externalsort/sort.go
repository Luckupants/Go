//go:build !solution

package externalsort

import (
	"container/heap"
	"io"
	"os"
	"sort"
	"strings"
)

type NewLineReader struct {
	bufSize int
	buffer  []byte
	reader  io.Reader
}

func (nlr *NewLineReader) Seek() (string, bool) {
	if len(nlr.buffer) == 0 {
		return "", false
	}
	answer := ""
	for i, r := range string(nlr.buffer) {
		if r == '\n' {
			answer = string(nlr.buffer)[:i]
			copy(nlr.buffer, string(nlr.buffer)[i+1:])
			nlr.buffer = nlr.buffer[:len(nlr.buffer)-len(string(nlr.buffer)[:i+1])]
			return answer, true
		}
	}
	answer = string(nlr.buffer)
	nlr.buffer = nlr.buffer[:0]
	return answer, false
}

func (nlr *NewLineReader) ReadLine() (string, error) {
	answer := strings.Builder{}
	str, found := nlr.Seek()
	answer.WriteString(str)
	if found {
		return answer.String(), nil
	}
	for !found {
		if len(nlr.buffer) == 0 {
			nlr.buffer = nlr.buffer[:nlr.bufSize]
			n, err := nlr.reader.Read(nlr.buffer)
			nlr.buffer = nlr.buffer[:n]
			if err == io.EOF {
				answer.WriteString(string(nlr.buffer))
				return answer.String(), err
			} else if err != nil {
				return "", err
			}
		}
		str, found = nlr.Seek()
		answer.WriteString(str)
	}
	return answer.String(), nil
}

func NewReader(r io.Reader) LineReader {
	answer := NewLineReader{bufSize: 1024, buffer: make([]byte, 1024), reader: r}
	answer.buffer = answer.buffer[:0]
	return &answer
}

type NewLineWriter struct {
	writer io.Writer
}

func NewWriter(w io.Writer) LineWriter {
	return &NewLineWriter{writer: w}
}

func (nlw *NewLineWriter) Write(l string) error {
	_, err := nlw.writer.Write([]byte(l + "\n"))
	return err
}

type Node struct {
	value string
	index int
}

type IntHeap []Node

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i].value < h[j].value }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x any) {
	*h = append(*h, x.(Node))
}

func (h *IntHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func ReadNumber(reader LineReader, index int, intHeap *IntHeap) error {
	str, err := reader.ReadLine()
	if err == nil || (err == io.EOF && len(str) > 0) {
		heap.Push(intHeap, Node{str, index})
	}
	return err
}

func Merge(w LineWriter, readers ...LineReader) error {
	hp := &IntHeap{}
	heap.Init(hp)
	for i, r := range readers {
		err := ReadNumber(r, i, hp)
		if err != nil && err != io.EOF {
			return err
		}
	}
	for len(*hp) != 0 {
		node := heap.Pop(hp)
		e := w.Write(node.(Node).value)
		if e != nil && e != io.EOF {
			return e
		}
		err := ReadNumber(readers[node.(Node).index], node.(Node).index, hp)
		if err != nil && err != io.EOF {
			return err
		}
	}
	return nil
}

func Sort(w io.Writer, in ...string) error {
	for _, path := range in {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rows := strings.Split(string(data), "\n")
		if len(rows) > 0 && rows[len(rows)-1] == "" {
			rows = rows[:len(rows)-1]
		}
		sort.Slice(rows, func(i int, j int) bool {
			return rows[i] < rows[j]
		})
		toWrite := strings.Builder{}
		for _, row := range rows {
			toWrite.WriteString(row + "\n")
		}
		err = os.WriteFile(path, []byte(toWrite.String()), 0777)
		if err != nil {
			return err
		}
	}
	files := make([]LineReader, len(in))
	for i, path := range in {
		f, err := os.Open(path)
		files[i] = NewReader(f)
		if err != nil {
			return err
		}
	}
	err := Merge(NewWriter(w), files...)
	return err
}
