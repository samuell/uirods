package main

import (
	"bytes"
	"fmt"
	"github.com/samuell/glow"
	"net/http"
)

func irodsPathHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>iRODS Thin UI</h1><p>Welcome user %s!</p>", r.URL.Path[1:])

	cmdIn := make(chan string, 0)
	linesOut := make(chan []byte, 16)

	glow.NewCommandExecutor(cmdIn, linesOut)
	cmdIn <- "ils"

	cnt := 0
	for line := range linesOut {
		if cnt > 0 {
			line := bytes.Replace(line, []byte("  C- "), []byte(""), 1)
			pathParts := bytes.Split(line, []byte("/"))
			folderName := pathParts[len(pathParts)-1]
			fmt.Fprint(w, "<li><a href=\"/irods", string(line), "\">", string(folderName), "</a></li>")
		} else {
			fmt.Fprint(w, "<p>Current folder: ", string(line), "</p>")
			fmt.Fprint(w, "<ul>")
		}
		cnt++
	}
	fmt.Fprint(w, "</ul>")
}

func main() {
	// Handle iRODS paths (for showing metadata etc)
	http.HandleFunc("/irods/", irodsPathHandler)
	// Handle file system paths
	http.Handle("/", http.FileServer(http.Dir(".")))
	// Serve on port 8080
	http.ListenAndServe(":8080", nil)
}
