//go:build !solution

package main

import (
	"errors"
	"flag"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"strconv"
	"time"
)

type HandlerError struct {
	statusCode int
	error
}

func digitToString(digit int) string {
	switch digit {
	case 0:
		return Zero
	case 1:
		return One
	case 2:
		return Two
	case 3:
		return Three
	case 4:
		return Four
	case 5:
		return Five
	case 6:
		return Six
	case 7:
		return Seven
	case 8:
		return Eight
	case 9:
		return Nine
	}
	return ""
}

func displayTime(timeToDisplay string, k int) (*image.RGBA, error) {
	digitWidth, columnWidth := 8*k, 4*k
	height := 12 * k
	ans := image.NewRGBA(image.Rect(0, 0, k*(6*8+2*4), k*12))
	curx := 0
	for _, letter := range timeToDisplay {
		var width int
		var strRepresentation string
		if letter == ':' {
			strRepresentation = Colon
			width = columnWidth
		} else {
			digit, err := strconv.Atoi(string(letter))
			if err != nil {
				return nil, err
			}
			strRepresentation = digitToString(digit)
			width = digitWidth
		}
		startx := curx
		endx := curx + width
		strIdx := 0
		colors := map[rune]color.RGBA{'.': {R: 255, G: 255, B: 255, A: 255}, '1': Cyan}
		str := []rune(strRepresentation)
		for y := 0; y < height; y += k {
			for x := startx; x < endx; x += k {
				for xx := x; xx < x+k; xx++ {
					for yy := y; yy < y+k; yy++ {
						ans.Set(xx, yy, colors[str[strIdx]])
					}
				}
				strIdx++
			}
			strIdx++
		}
		curx = endx
	}
	return ans, nil
}

func main() {
	port := flag.String("port", "8000", "port to start server on")
	flag.Parse()
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
	getHandler := func(w http.ResponseWriter, r *http.Request) ([]byte, error) {
		timeToDisplay := r.URL.Query().Get("time")
		if timeToDisplay == "" {
			timeToDisplay = time.Now().Format("15:04:05")
		}
		var err error
		_, err = time.Parse("15:04:05", timeToDisplay)
		if err != nil || len(timeToDisplay) != 8 {
			return nil, HandlerError{statusCode: http.StatusBadRequest, error: err}
		}
		kk := r.URL.Query().Get("k")
		var k int
		if kk == "" {
			k = 1
		} else {
			k, err = strconv.Atoi(kk)
			if err != nil {
				return nil, HandlerError{statusCode: http.StatusBadRequest, error: err}
			}
		}
		if k < 1 || k > 30 {
			return nil, HandlerError{statusCode: http.StatusBadRequest, error: errors.New("k cannot be less than 1 or greater than 30")}
		}
		forImage, err := displayTime(timeToDisplay, k)
		if err != nil {
			return nil, HandlerError{statusCode: http.StatusBadRequest, error: err}
		}
		w.Header().Set("Content-Type", "image/png")
		_ = png.Encode(w, forImage)
		return nil, nil
	}
	server.HandleFunc("GET /", wrapErrorReply(getHandler))
	_ = http.ListenAndServe(":"+*port, server)
}
