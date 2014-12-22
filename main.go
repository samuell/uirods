package main

import (
	"fmt"
	"github.com/samuell/glow"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func irodsPathHandler(w http.ResponseWriter, r *http.Request) {
	irodsMntPath := os.Getenv("IRODSMNT_IRODSPATH")

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
	cwd := ""
	for line := range linesOut {
		var isFolder bool
		line := string(line)
		if strings.Contains(line, "  C- ") {
			line = strings.Replace(line, "  C- ", "", 1)
			isFolder = true
		} else {
			line = strings.Replace(line, " ", "", 1)
			isFolder = false
		}
		pathParts := strings.Split(line, "/")
		if cnt > 0 {
			fileName := pathParts[len(pathParts)-1]
			if isFolder {
				fmt.Fprint(w, "<li><a href=\"/irods", string(line), "\">", string(fileName), "</a></li>")
			} else {
				line = strings.Replace(line, " ", "", 1)
				var cwdLocal string
				if cwd == "/" {
					cwdLocal = cwd
				} else {
					cwdLocal = strings.Replace(cwd, irodsMntPath, "", 1)
				}
				fmt.Fprint(w, "<li><a href=\"/files/", cwdLocal, "/", line, "\">", string(fileName), "</a></li>")
			}
		} else {
			cwd = strings.Replace(line, ":", "", 1)
			fmt.Fprint(w, "<p>Current folder: ", cwd, "</p>")
			fmt.Fprint(w, "<p><a href=\"/irods", strings.Join(pathParts[:len(pathParts)-1], "/"), "\">Parent folder</a></p>")
			fmt.Fprint(w, "<ul>")
		}
		cnt++
	}
	fmt.Fprint(w, "</ul>")
}

func main() {
	physicalMntPath := os.Getenv("IRODSMNT_PHYSPATH")

	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(physicalMntPath))))
	http.HandleFunc("/irods/", irodsPathHandler)
	http.HandleFunc("/", irodsPathHandler)
	http.ListenAndServe(":9595", nil)
}
