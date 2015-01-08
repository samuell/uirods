#!/bin/bash
export IRODSMNT_IRODSPATH=/
export IRODSMNT_FILESPATH=~/mnt/irods
icd /
irodsFs ~/mnt/irods
go run main.go -p 8081
