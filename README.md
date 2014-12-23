uiRODS - Simple Web UI for iRODS
====

![uiRODS Screenshot](uirods_screenshot.png?raw=true)

# Prerequisites

* Go 1.x
* A properly configured ```$GOPATH``` environment variable
* The iRODS FUSE module (comes installed, with the icommands .deb package from irods.org)

# Installation

1. Get the code:
````bash
go get github.com/samuell/uirods
````

2. Build it in some suitable folder, eg:
````bash
mkdir -p ~/opt/uirods
cd ~/opt/uirods
go build github.com/samuell/uirods
````

3. Mount an irods folder in a local folder, using iRODS FUSE:
````bash
mkdir -p ~/mnt/irods
iinit
icd /
irodsFs ~/mnt/irods
````

4. Set up some environmental variables (add this to your ```~/.bashrc``` or ```~/.bash_profile``` to make it last longer than the current bash session):
````bash
export IRODSMNT_FILESPATH='~/mnt/irods'
export IRODSMNT_IRODSPATH='/' # If you did icd / above, before mounting!
````

5. Start the web server:
````bash
./uirods
````
6. Surf in to [http://localhost:8080](http://localhost:8080) in your browser!
7. Done, enjoy your new (but most probably highly insecure) iRODS ui! :)
