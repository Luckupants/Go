//go:build !solution

package httpgauge

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Gauge struct {
	mx      sync.Mutex
	counter map[string]int
	router  chi.Router
}

func New() *Gauge {
	ans := &Gauge{counter: make(map[string]int), router: chi.NewRouter()}
	ans.router.Get("/user/{userID}", func(w http.ResponseWriter, r *http.Request) {
		ans.mx.Lock()
		defer ans.mx.Unlock()
		ans.counter["/user/{userID}"]++
	})
	ans.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		ans.mx.Lock()
		defer ans.mx.Unlock()
		ans.counter[r.URL.Path]++
	})
	return ans
}

func (g *Gauge) Snapshot() map[string]int {
	return g.counter
}

// ServeHTTP returns accumulated statistics in text format ordered by pattern.
//
// For example:
//
//	/a 10
//	/b 5
//	/c/{id} 7
func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = r
	var toWrite strings.Builder
	var paths []string
	for path := range g.counter {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		count := strconv.Itoa(g.counter[path])
		toWrite.WriteString(fmt.Sprintf("%s %s\n", path, count))
	}
	_, _ = w.Write([]byte(toWrite.String()))
}

func (g *Gauge) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.router.ServeHTTP(w, r)
		next.ServeHTTP(w, r)
	})
}
