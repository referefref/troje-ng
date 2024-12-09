#!/bin/bash
set -e

# Build Troje
go build -o troje ./cmd/troje

# Start Troje
sudo ./troje -b troje_base -l :8022 -idle 1h
