#!/bin/bash

mkdir distribution

sudo apt-get update
sudo apt-get install -y libgtk-3-dev

go get github.com/GeertJohan/go.rice
go get github.com/GeertJohan/go.rice/rice

go build
rice append --exec D3pixelbot

wget https://github.com/c-smile/sciter-sdk/raw/master/bin.gtk/x64/libsciter-gtk.so

# 7z a -t7z distribution/Linux.x86-64.7z -m0=lzma2 -mx=9 -aoa D3pixelbot README.md LICENSE config.json libsciter-gtk.so
env GZIP=-9 tar cfzv distribution/Linux.x86-64.tar.gz D3pixelbot README.md LICENSE config.json libsciter-gtk.so

rm D3pixelbot
rm libsciter-gtk.so