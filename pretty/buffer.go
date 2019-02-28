package pretty

import (
	"bytes"
	"io"
)

type bf struct {
	n int
	f func(io.Writer, string) error
}

type Buffer struct {
	bytes.Buffer
	bfs []bf
	n   int
}

func (b *Buffer) Ugly() { b.appendf(nil) }

func (b *Buffer) Format(w io.Writer) error {
	if len(b.bfs) == 0 {
		b.bfs = []bf{bf{}}
	}
	b.appendf(nil)
	for _, bf := range b.bfs {
		switch {
		case bf.f == nil:
			if _, err := w.Write(b.Next(bf.n)); err != nil {
				return err
			}
		default:
			if err := bf.f(w, string(b.Next(bf.n))); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Buffer) Reset() {
	b.Buffer.Reset()
	b.bfs = []bf{bf{}}
	b.n = 0
}

func (b *Buffer) appendf(f func(io.Writer, string) error) {
	if l := len(b.bfs); l > 0 {
		b.bfs[l-1].n = b.Len() - b.n
		b.n += b.bfs[l-1].n
	}
	b.bfs = append(b.bfs, bf{f: f})
}
