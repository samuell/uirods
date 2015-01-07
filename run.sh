#!/bin/bash
export IRODSMNT_IRODSPATH="/"
export IRODSMNT_FILESPATH="~/mnt/irods"
go build github.com/samuell/uirods
./uirods
