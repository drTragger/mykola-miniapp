package main

import (
	"log"
	"net/http"
	"os"

	"github.com/drTragger/mykola-miniapp/webui"
)

func main() {
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8090"
	}

	handler, err := webui.NewHandler()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Mini App server started on", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
