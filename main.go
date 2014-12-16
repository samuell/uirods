package main

import (
	"fmt"
	"github.com/samuell/glow"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>iRODS Thin UI</h1><p>Welcome user %s!</p>", r.URL.Path[1:])
	cmdIn := make(chan string, 0)
	linesOut := make(chan []byte, 16)
	glow.NewCommandExecutor(cmdIn, linesOut)
	cmdIn <- "ils"
	fmt.Fprint(w, "<ul>")
	for line := range linesOut {
		fmt.Fprint(w, "<li>", string(line), "</li>")
	}
	fmt.Fprint(w, "</ul>")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
