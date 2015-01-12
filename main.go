package main

import (
	"flag"
	"fmt"
	"github.com/samuell/glow"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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

	headerHtml = `<html>
	<head>
	<title>uiRODS</title>
	<style>
	body{font-family:arial,helvetica,sans-serif;}
	.cwd{background:#efefef;color:#777;padding:.2em .4em;}
	table th,table td{border-width: 1px 0 0 1px;border-style: dotted;border-color: #ccc;padding: .4em .7em;}
	</style>
	</head>
	<body>
	<h1>uiRODS</h1>`
	footerHtml = "</body></html>"

	cwd string
	cnt int
)

// --------------------------------------------------------------------------------
// Main function
// --------------------------------------------------------------------------------

func main() {
	flag.Parse()

	fmt.Printf("Starting, assuming iRODS folder %s is mounted at %s ...\n", irodsMntPath, filesMntPath)

	http.Handle(fileServerBasePath+"/", http.StripPrefix(fileServerBasePath+"/", http.FileServer(http.Dir(filesMntPath))))
	http.HandleFunc(iRodsHandlerBasePath+"/", irodsPathHandler)
	http.HandleFunc(fileHandlerBasePath+"/", irodsFileHandler)
	http.HandleFunc("/", indexHandler)

	bind := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Listening on http://%v", bind)
	log.Fatal("FATAL: %v", http.ListenAndServe(bind, nil))
}

// --------------------------------------------------------------------------------
// Handlers
// --------------------------------------------------------------------------------

// Handle the root url / index page.
// Show a link to start browsing the iRODS folder tree.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, headerHtml)
	fmt.Fprintf(w, "<ul><li><a href=\"%s/tempZone\">Open uiRods browser</a></li></ul>", iRodsHandlerBasePath)
	fmt.Fprint(w, footerHtml)
}

// Handle URLs representing iRODS folder paths
// Show links for navigating in the folder tree.
func irodsPathHandler(w http.ResponseWriter, r *http.Request) {
	// Output the header
	fmt.Fprint(w, headerHtml)

	// Change iRODS current folder to the requested path, using the icd command
	targetFolder := strings.Replace(r.URL.RequestURI(), iRodsHandlerBasePath, "", 1)
	cmdBinary := "icd"
	cmdParams := targetFolder
	cmd := exec.Command(cmdBinary, cmdParams)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error when executing command 'icd %s': %s", targetFolder, err)
	}

	// Execute the ils command, and iterate over the iput on the linesOut channel
	cmdIn := make(chan string, 0)
	linesOut := make(chan []byte, 16)
	glow.NewCommandExecutor(cmdIn, linesOut)
	cmdIn <- "ils"
	cnt = 0

	// Loop over lines in output from ils
	for line := range linesOut {
		var isFolder bool
		line := string(line)

		// Check if current line represents a folder or file
		if representsFolder(line) { // "C-" designates folders in ils output
			line = stripFolderMarker(line)
			isFolder = true
		} else {
			line = stripFirstSpace(line)
			isFolder = false
		}

		pathParts := strings.Split(line, "/")

		if cnt == 0 {
			// For the first line, which does not contain any paths anyway,
			// just rite some "introductory" HTML

			// Write current working directory
			cwd = strings.Replace(line, ":", "", 1)
			fmt.Fprint(w, "<p class=\"cwd\">Current folder: ", cwd, "</p>")

			// Write parent folder link
			parentFolderPath := strings.Join(pathParts[:len(pathParts)-1], "/")
			fmt.Fprintf(w, "<p><a href=\"%s%s\">&laquo; Parent folder</a></p>", iRodsHandlerBasePath, parentFolderPath)

			// Start the file/folder list
			fmt.Fprint(w, "<ul>")
		} else {
			// The rest of the ils output lines

			fileName := pathParts[len(pathParts)-1]

			// Print the actual folder / file links (as list items)
			if isFolder {
				if fileName != "" {
					fmt.Fprint(w, "<li><a href=\"", iRodsHandlerBasePath, line, "\">", fileName, "</a></li>")
				}
			} else {
				line = strings.Replace(line, " ", "", 1)
				var cwdLocal string
				if cwd == "/" {
					cwdLocal = cwd
				} else {
					cwdLocal = strings.Replace(cwd, irodsMntPath, "", 1)
				}
				fmt.Fprintf(w, "<li><a href=\"%s/%s/%s\">%s</a></li>", fileHandlerBasePath, cwdLocal, line, fileName)
			}
		}
		cnt++
	}
	fmt.Fprint(w, "</ul>")
	fmt.Fprint(w, footerHtml)
}

// Handle URLS representing iRODS file paths.
// Show metadata and download link
func irodsFileHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, headerHtml)

	// Extract the iRODS file path from the URL
	filePath := strings.Replace(r.URL.RequestURI(), fileHandlerBasePath, "", 1)
	fileParts := strings.Split(filePath, "/")
	parentFolderPath := strings.Join(fileParts[:len(fileParts)-1], "/")

	// Print HTML content for presenting the file
	fmt.Fprintf(w, "<p class=\"cwd\">Current file: %s</p>", filePath)
	fmt.Fprintf(w, "<p><a href=\"%s/%s\">&laquo; Parent folder</a></p>", iRodsHandlerBasePath, parentFolderPath)
	fmt.Fprint(w, "<h4>Download file</h4>")
	fmt.Fprintf(w, "<ul><li><a href=\"%s/%s\">%s</a></li></ul>", fileServerBasePath, filePath, filePath)
	fmt.Fprint(w, "<h4>Metadata</h4>")
	fmt.Fprint(w, "<table><tr><th>Attribute</th><th>Value</th><th>Units</th></tr>")

	// Get the metadata about the current file
	cmdOut, cmdErr := exec.Command("imeta", "ls", "-d", filePath).Output()
	if cmdErr != nil {
		log.Fatal("Failed executing imeta command for ", filePath, ": ", cmdErr)
	}
	metaStr := stripFirstLine(string(cmdOut))

	// Split meta data triplets into chunks, with one triplet in each chunk
	metaChunks := strings.Split(metaStr, "----")
	// Loop over meta data "triplets" or "chunks"
	for _, metaChunk := range metaChunks {
		// Extract attribute [name]
		attr := getMetaDataFieldValue("attribute", metaChunk)
		value := getMetaDataFieldValue("value", metaChunk)
		units := getMetaDataFieldValue("units", metaChunk)

		// Print a table row with attribute name, value and units
		fmt.Fprintf(w, "<tr style=\"border-bottom: 1px solid grey;\"><td>%s</td><td>%s</td><td>%s</td></tr>", attr, value, units)
	}
	fmt.Fprint(w, "</table>")
	fmt.Fprint(w, footerHtml)
}

// --------------------------------------------------------------------------------
// Helper functions
// --------------------------------------------------------------------------------

func getMetaDataFieldValue(fieldName string, metaData string) string {
	var value string
	pat, _ := regexp.Compile(fieldName + ": ([a-zA-Z0-9]+)")
	matches := pat.FindStringSubmatch(metaData)
	if len(matches) > 0 {
		value = matches[1]
	}
	return value
}

func stripFirstLine(str string) string {
	lines := strings.Split(str, "\n")
	newLines := lines[1:]
	newText := strings.Join(newLines, "\n")
	return newText
}

func representsFolder(line string) bool {
	return strings.Contains(line, "  C- ")
}

func stripFolderMarker(line string) string {
	return strings.Replace(line, "  C- ", "", 1)
}

func stripFirstSpace(line string) string {
	return strings.Replace(line, " ", "", 1)
}
