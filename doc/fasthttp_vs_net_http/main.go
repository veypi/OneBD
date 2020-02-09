package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
)

var response = []byte("response")

func NetHttpServer() {
	fmt.Printf("net http\n")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(response)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Print(err)
	}
}

func FastHttpServer() {
	fmt.Printf("fast http\n")
	err := fasthttp.ListenAndServe(":8080", func(ctx *fasthttp.RequestCtx) {
		ctx.Write(response)
	})
	if err != nil {
		fmt.Print(err)
	}
}

func main() {
	//NetHttpServer()
	FastHttpServer()
}
