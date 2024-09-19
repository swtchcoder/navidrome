package slice

import (
	"bufio"
	"bytes"
	"io"
	"iter"
)

func Map[T any, R any](t []T, mapFunc func(T) R) []R {
	r := make([]R, len(t))
	for i, e := range t {
		r[i] = mapFunc(e)
	}
	return r
}

func Group[T any, K comparable](s []T, keyFunc func(T) K) map[K][]T {
	m := map[K][]T{}
	for _, item := range s {
		k := keyFunc(item)
		m[k] = append(m[k], item)
	}
	return m
}

func MostFrequent[T comparable](list []T) T {
	if len(list) == 0 {
		var zero T
		return zero
	}
	var topItem T
	var topCount int
	counters := map[T]int{}

	if len(list) == 1 {
		topItem = list[0]
	} else {
		for _, id := range list {
			c := counters[id] + 1
			counters[id] = c
			if c > topCount {
				topItem = id
				topCount = c
			}
		}
	}

	return topItem
}

func Insert[T any](slice []T, value T, index int) []T {
	return append(slice[:index], append([]T{value}, slice[index:]...)...)
}

func Remove[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}

func Move[T any](slice []T, srcIndex int, dstIndex int) []T {
	value := slice[srcIndex]
	return Insert(Remove(slice, srcIndex), value, dstIndex)
}

func BreakUp[T any](items []T, chunkSize int) [][]T {
	numTracks := len(items)
	var chunks [][]T
	for i := 0; i < numTracks; i += chunkSize {
		end := i + chunkSize
		if end > numTracks {
			end = numTracks
		}

		chunks = append(chunks, items[i:end])
	}
	return chunks
}

func RangeByChunks[T any](items []T, chunkSize int, cb func([]T) error) error {
	chunks := BreakUp(items, chunkSize)
	for _, chunk := range chunks {
		err := cb(chunk)
		if err != nil {
			return err
		}
	}
	return nil
}

func LinesFrom(reader io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		scanner := bufio.NewScanner(reader)
		scanner.Split(scanLines)
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				return
			}
		}
	}
}

// From https://stackoverflow.com/a/41433698
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		if data[i] == '\n' {
			// We have a line terminated by single newline.
			return i + 1, data[0:i], nil
		}
		advance = i + 1
		if len(data) > i+1 && data[i+1] == '\n' {
			advance += 1
		}
		return advance, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func CollectChunks[T any](n int, it iter.Seq[T]) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		var s []T
		for x := range it {
			s = append(s, x)
			if len(s) >= n {
				if !yield(s) {
					return
				}
				s = nil
			}
		}
		if len(s) > 0 {
			yield(s)
		}
	}
}
