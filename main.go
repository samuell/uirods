package main

import (
	"bytes"
	"fmt"
	"github.com/samuell/glow"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func irodsPathHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>uiRODS</h1>")

	targetFolder := strings.Replace(r.URL.RequestURI(), "/irods", "", 1)

	cmd := exec.Command("icd", targetFolder)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmdIn := make(chan string, 0)
	linesOut := make(chan []byte, 16)

	glow.NewCommandExecutor(cmdIn, linesOut)
	cmdIn <- "ils"

	cnt := 0
	for line := range linesOut {
		line := bytes.Replace(line, []byte("  C- "), []byte(""), 1)
		pathParts := bytes.Split(line, []byte("/"))
		if cnt > 0 {
			folderName := pathParts[len(pathParts)-1]
			fmt.Fprint(w, "<li><a href=\"/irods", string(line), "\">", string(folderName), "</a></li>")
		} else {
			fmt.Fprint(w, "<p>Current folder: ", string(line), "</p>")
			fmt.Fprint(w, "<p><a href=\"/irods", string(bytes.Join(pathParts[:len(pathParts)-1], []byte("/"))), "\">Parent folder</a></p>")
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
