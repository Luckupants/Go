//go:build !solution

package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"reflect"
	"time"
)

type HandlerError struct {
	statusCode int
	error
}

func WrapErrorReply(h func(w http.ResponseWriter, r *http.Request) ([]byte, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err HandlerError
		toWrite, er := h(w, r)
		if er != nil {
			errors.As(er, &err)
			if err.error != nil {
				w.Header().Set("error", err.Error())
			}
			w.WriteHeader(err.statusCode)
		}
		_, _ = w.Write(toWrite)
	}
}

func MakeHandler(service interface{}) http.Handler {
	server := http.NewServeMux()
	serviceType := reflect.TypeOf(service)
	methodAmnt := serviceType.NumMethod()
	var neededMethods []int
	for i := 0; i < methodAmnt; i++ {
		if serviceType.Method(i).Type.NumIn() == 3 && serviceType.Method(i).Type.In(1).String() == "context.Context" && serviceType.Method(i).Type.In(2).Kind() == reflect.Pointer && serviceType.Method(i).Type.NumOut() == 2 {
			neededMethods = append(neededMethods, i)
		}
	}
	for _, idx := range neededMethods {
		method := serviceType.Method(idx)
		handler := func(w http.ResponseWriter, r *http.Request) ([]byte, error) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, HandlerError{statusCode: http.StatusInternalServerError, error: err}
			}
			reqType := method.Type.In(2).Elem()
			reqObj := reflect.New(reqType).Interface()
			err = json.Unmarshal(body, reqObj)
			if err != nil {
				return nil, HandlerError{statusCode: http.StatusInternalServerError, error: err}
			}
			result := method.Func.Call([]reflect.Value{reflect.ValueOf(service), reflect.ValueOf(r.Context()), reflect.ValueOf(reqObj)})
			if len(result) != 2 {
				return nil, HandlerError{statusCode: http.StatusInternalServerError, error: errors.New("Call return argument count mismatch")}
			}
			if result[1].Interface() != any(error(nil)) {
				return nil, HandlerError{statusCode: http.StatusInternalServerError, error: errors.New(fmt.Sprint(result[1].Interface()))}
			}
			resObj, err := json.Marshal(result[0].Interface())
			if err != nil {
				return nil, HandlerError{statusCode: http.StatusInternalServerError, error: err}
			}
			if err != nil {
				return nil, HandlerError{statusCode: http.StatusInternalServerError, error: err}
			}
			return resObj, nil
		}
		server.HandleFunc(method.Name+" /", WrapErrorReply(handler))
	}
	return server
}

func Call(ctx context.Context, endpoint string, method string, req, rsp any) error {
	c := &http.Client{Timeout: time.Second * 10}
	reqObg, err := json.Marshal(req)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(reqObg))
	if err != nil {
		return err
	}
	resp, err := c.Do(request)
	fmt.Println(resp.Header.Get("error"))
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.Header.Get("error"))
		return errors.New(resp.Header.Get("error"))
	}
	respObj := reflect.New(reflect.TypeOf(rsp).Elem()).Interface()
	err = json.Unmarshal(body, respObj)
	if err != nil {
		return err
	}
	reflect.ValueOf(rsp).Elem().Set(reflect.ValueOf(respObj).Elem())
	return nil
}
