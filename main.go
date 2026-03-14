package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func serveContent(w http.ResponseWriter, req *http.Request) {
	rst := req.URL.Path
	file_name := strings.TrimPrefix(rst, "/") + ".html"
	file, err := os.ReadFile(file_name)
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	fmt.Fprintf(w, "%s\n", file)
}

func verb(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	switch method {
	case "GET":
		fmt.Fprintf(w, "yoyoGET")
	case "POST":
		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		fmt.Fprintf(w, "You sent: %s\n", body)
	case "PUT":
		fmt.Fprintf(w, "yoyoPUT")
	case "DELETE":
		fmt.Fprintf(w, "yoyoDELETE")
	case "PATCH":
		fmt.Fprintf(w, "yoyoPATCH")
	case "HEAD":
		fmt.Fprintf(w, "yoyoHEAD")
	default:
		fmt.Fprintf(w, "Method not allowed")
	}
}

// func print_headers(w http.ResponseWriter,req *http.Request) {
// 	for name, headers := range req.Header {
// 		for _, h := range headers {
// 			fmt.Fprintf(w, "%v: %v\n", name, h)
// 		}
// 	}
// }

func main() {
	// http.HandleFunc("/hello", hello)
	// http.HandleFunc("/print_headers", print_headers)
	http.HandleFunc("/", serveContent)
	http.HandleFunc("/api", verb)
	// http.Handle("/", http.FileServer(http.Dir("./")))

	http.ListenAndServe(":8090", nil)
}
