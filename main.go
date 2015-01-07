package main

import (
	"flag"
	"fmt"
	"github.com/samuell/glow"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	iRodsHandlerBasePath = "/irods"
	fileServerBasePath   = "/files"
	fileHandlerBasePath  = "/file"
)

var (
	port = flag.Int("p", 8080, "HTTP Listening port")
	host = flag.String("h", "localhost", "HTTP Listening host")

	filesMntPath = os.Getenv("IRODSMNT_FILESPATH")
	irodsMntPath = os.Getenv("IRODSMNT_IRODSPATH")

	headerHtml = `<html><head><title>uiRODS</title>
	<style>body{font-family:arial,helvetica,sans-serif;}.cwd{background:#efefef;color:#777;padding:.2em .4em;}</style>
	</head><body><h1>uiRODS</h1>`
	footerHtml = "</body></html>"

	cwd string
	cnt int
)

// --------------------------------------------------------------------------------
// Handlers
// --------------------------------------------------------------------------------

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, headerHtml)
	fmt.Fprint(w, "<ul><li><a href=\"", iRodsHandlerBasePath, "/tempZone\">Open uiRods browser</a></li></ul>")
	fmt.Fprint(w, footerHtml)
}

func irodsPathHandler(w http.ResponseWriter, r *http.Request) {
	// Output the header
	fmt.Fprint(w, headerHtml)

	// Change iRODS current folder to the requested path, using the icd command
	targetFolder := strings.Replace(r.URL.RequestURI(), iRodsHandlerBasePath, "", 1)
	cmdBinary := "icd"
	cmdParams := targetFolder
	cmd := exec.Command(cmdBinary, cmdParams)
	//fmt.Println("Now executing command: ", cmdBinary, " ", cmdParams)
	err := cmd.Run()
	if err != nil {
		log.Fatal("Error when executing command 'icd ", targetFolder, "': ", err)
	}

	// Execute the ils command, and iterate over the iput on the linesOut channel
	cmdIn := make(chan string, 0)
	linesOut := make(chan []byte, 16)
	glow.NewCommandExecutor(cmdIn, linesOut)
	cmdIn <- "ils"
	cnt = 0
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
				fmt.Fprint(w, "<li><a href=\"", iRodsHandlerBasePath, string(line), "\">", string(fileName), "</a></li>")
			} else {
				line = strings.Replace(line, " ", "", 1)
				var cwdLocal string
				if cwd == "/" {
					cwdLocal = cwd
				} else {
					cwdLocal = strings.Replace(cwd, irodsMntPath, "", 1)
				}
				fmt.Fprint(w, "<li><a href=\"", fileHandlerBasePath, "/", cwdLocal, "/", line, "\">", string(fileName), "</a></li>")
			}
		} else {
			cwd = strings.Replace(line, ":", "", 1)
			fmt.Fprint(w, "<p class=\"cwd\">Current folder: ", cwd, "</p>")
			fmt.Fprint(w, "<p><a href=\"", iRodsHandlerBasePath, strings.Join(pathParts[:len(pathParts)-1], "/"), "\">Parent folder</a></p>")
			fmt.Fprint(w, "<ul>")
		}
		cnt++
	}
	fmt.Fprint(w, "</ul>")
	fmt.Fprint(w, footerHtml)
}

func irodsFileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, headerHtml)
	filePath := strings.Replace(r.URL.RequestURI(), fileHandlerBasePath, "", 1)
	fmt.Fprint(w, "<ul><li>File: ", filePath, "</li></ul>")
	fmt.Fprint(w, footerHtml)
}

// --------------------------------------------------------------------------------
// Main function
// --------------------------------------------------------------------------------

func main() {
	flag.Parse()

	http.Handle(fileServerBasePath+"/", http.StripPrefix(fileServerBasePath+"/", http.FileServer(http.Dir(filesMntPath))))
	http.HandleFunc(iRodsHandlerBasePath+"/", irodsPathHandler)
	http.HandleFunc(fileHandlerBasePath+"/", irodsFileHandler)
	http.HandleFunc("/", indexHandler)

	bind := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Listening on http://%v", bind)
	log.Fatal("FATAL: %v", http.ListenAndServe(bind, nil))
}
