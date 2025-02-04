#!/bin/sh

cat > main.cpp

clang++ main.cpp -o main
./main

rm main main.cpp