#! /usr/bin/bash
rm -r release/*
cd web-service
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o audio-vault.exe *.go
mv audio-vault.exe ../release/audio-vault.exe
cp audio-vault.db ../release/audio-vault.db
cd ..
mkdir release/static-assets/
mkdir release/views/
cp -R ./web-service/static-assets/* ./release/static-assets/
cp -R ./web-service/views/* ./release/views/
zip -9 -r audio-vault.zip release
mv audio-vault.zip ./release/
