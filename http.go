package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func GET(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			ResponseError(res, http.StatusMethodNotAllowed, nil)
			return
		}
		next.ServeHTTP(res, req)
	})
}

func ResponseError(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	io.WriteString(w, http.StatusText(statusCode))
}

func ResponseResult(writer http.ResponseWriter, response Response) {
	buf, err := json.Marshal(response)
	if err != nil {
		ResponseError(writer, 500, err)
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(writer, string(buf))
	log.Print(string(buf))
}
