//go:build !solution

package genericsum

import (
	"golang.org/x/exp/constraints"
	"math/cmplx"
	"sync/atomic"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func SortSlice[T constraints.Ordered](a []T) {
	for i := 0; i < len(a); i++ {
		for j := i + 1; j < len(a); j++ {
			if a[i] > a[j] {
				a[i], a[j] = a[j], a[i]
			}
		}

	}
}

func MapsEqual[K, V comparable](a, b map[K]V) bool {
	for key, value := range a {
		if _, ok := b[key]; !ok || b[key] != value {
			return false
		}
	}
	for key, value := range b {
		if _, ok := a[key]; !ok || a[key] != value {
			return false
		}
	}
	return true
}

func SliceContains[T comparable](s []T, v T) bool {
	for _, value := range s {
		if value == v {
			return true
		}
	}
	return false
}

func MergeChans[T any](chs ...<-chan T) <-chan T {
	res := make(chan T)
	cnt := atomic.Int32{}
	cnt.Add(int32(len(chs)))
	for _, ch := range chs {
		ch := ch
		go func() {
			for msg := range ch {
				res <- msg
			}
			if cnt.Add(-1) == 0 {
				close(res)
			}
		}()
	}
	return res
}

func Sopryazheni[T constraints.Complex | constraints.Integer | constraints.Float](a, b T) bool {
	switch v, g := any(a), any(b); v.(type) {
	case complex64:
		aa, bb := v.(complex64), g.(complex64)
		return cmplx.Conj(complex128(aa)) == complex128(bb)
	case complex128:
		aa, bb := v.(complex128), g.(complex128)
		return cmplx.Conj(aa) == bb
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return v == g
	}
	return false // should be panic?
}

func IsHermitianMatrix[T constraints.Complex | constraints.Integer | constraints.Float](m [][]T) bool {
	if len(m) == 0 {
		return true
	}
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[0]); j++ {
			if !Sopryazheni(m[i][j], m[j][i]) {
				return false
			}
		}
	}
	return true
}
