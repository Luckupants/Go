//go:build !solution

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const keyGenerateTries = 1000

type ShortenedURL struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

type PostBody struct {
	URL string `json:"url"`
}

type HandlerError struct {
	statusCode int
	error
}

func main() {
	port := flag.String("port", "8000", "port to start server on")
	flag.Parse()
	urlToKeys := make(map[string]string)
	keyToUrls := make(map[string]string)
	var mx sync.Mutex
	server := http.NewServeMux()
	wrapErrorReply := func(h func(w http.ResponseWriter, r *http.Request) ([]byte, error)) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			var err HandlerError
			toWrite, er := h(w, r)
			if er != nil {
				errors.As(er, &err)
				w.WriteHeader(err.statusCode)
				if err.error != nil {
					w.Header().Set("error", err.Error())
				}
			}
			_, _ = w.Write(toWrite)
		}
	}
	postHandler := func(w http.ResponseWriter, r *http.Request) ([]byte, error) {
		var body []byte
		var err error
		body, err = io.ReadAll(r.Body)
		// err = r.Body.Close()   Написано, что необязательно вроде
		if err != nil {
			return nil, HandlerError{http.StatusInternalServerError, err}
		}
		var postBody PostBody
		err = json.Unmarshal(body, &postBody)
		if err != nil {
			return nil, HandlerError{http.StatusBadRequest, err}
		}
		mx.Lock()
		ans, found := urlToKeys[postBody.URL]
		mx.Unlock()
		if found {
			answerJSON, err := json.Marshal(ShortenedURL{URL: postBody.URL, Key: ans})
			if err != nil {
				return nil, HandlerError{statusCode: http.StatusInternalServerError, error: err}
			}
			w.Header().Set("Content-Type", "application/json")
			return answerJSON, HandlerError{statusCode: http.StatusOK, error: errors.New("url already registered")}
		}
		var key string
		i := 0
		for ; i < keyGenerateTries; i++ {
			key = strconv.Itoa(rand.Intn(int(math.Pow10(6))))
			if _, found := keyToUrls[key]; !found {
				break
			}
		}
		if i == keyGenerateTries {
			return nil, HandlerError{statusCode: http.StatusContinue, error: errors.New("couldn't generate key, please retry")}
		}
		answer := ShortenedURL{URL: postBody.URL, Key: key}
		mx.Lock()
		keyToUrls[key] = postBody.URL
		urlToKeys[postBody.URL] = key
		mx.Unlock()
		answerJSON, err := json.Marshal(answer)
		if err != nil {
			return nil, HandlerError{http.StatusInternalServerError, err}
		}
		w.Header().Set("Content-Type", "application/json")
		return answerJSON, nil
	}
	server.HandleFunc("POST /shorten", wrapErrorReply(postHandler))
	getHandler := func(w http.ResponseWriter, r *http.Request) ([]byte, error) {
		key, found := strings.CutPrefix(r.URL.Path, "/go/")
		if !found {
			return nil, HandlerError{statusCode: http.StatusBadRequest, error: errors.New("something went wrong... wrong url")}
		}
		fmt.Println(r.URL.Path)
		mx.Lock()
		val, ok := keyToUrls[key]
		mx.Unlock()
		if !ok {
			return nil, HandlerError{statusCode: http.StatusNotFound, error: errors.New("wrong key : " + key)}
		}
		http.Redirect(w, r, val, http.StatusFound)
		return nil, nil
	}
	server.HandleFunc("GET /go/", wrapErrorReply(getHandler))
	_ = http.ListenAndServe(":"+*port, server)
}
