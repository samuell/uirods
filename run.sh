#!/bin/bash
export IRODSMNT_IRODSPATH=/
export IRODSMNT_FILESPATH=~/mnt/irods
icd $IRODSMNT_IRODSPATH
irodsFs $IRODSMNT_IRODSPATH
go run main.go -p 8081
