#!/bin/bash
echo "start gen proto"
read -p "Enter the plugins name: " plugins
read -p "Enter out path:" path

cd protobuf
protoc --go_out=plugins=${plugins}:. ${path}
