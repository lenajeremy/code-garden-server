#!/bin/sh

# Read the input Go source code and write to file.go
cat > file.go

## Seek to the start of the file
#exec 3<file.go
#exec 3<&-

# Run the Go source code
go run file.go

# Delete the Go source code file
rm file.go