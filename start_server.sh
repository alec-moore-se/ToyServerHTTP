#!/usr/bin/env bash

echo "Starting server"

go build -o ToyServerHTTP
./ToyServerHTTP
