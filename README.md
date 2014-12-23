uiRODS - Simple Web UI for iRODS
====

![uiRODS Screenshot](uirods_screenshot.png?raw=true)

# Prerequisit

* Go 1.x
* A properly configured $GOPATH environment variable

# Installation

Get the code:
````bash
go get github.com/samuell/uirods
````
Build it:
````bash
cd <to some folder>
go build github.com/samuell/uirods
````
Run:
````bash
./uirods -p <portno>
````
Now surf in to http://localhost:<portno> in your browser!

