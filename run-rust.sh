#!/bin/sh

cat > file.rs

rustc file.rs -o output
./output