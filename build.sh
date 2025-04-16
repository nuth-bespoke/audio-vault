#! /usr/bin/bash
rm -r release/*
cd web-service
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s -X main.GIT_COMMIT_HASH=`git rev-parse HEAD`" -o audio-vault.exe *.go
mv audio-vault.exe ../release/audio-vault.exe
#upx-ucl --best -o "../release/audio-vault.exe" audio-vault.exe
cp audio-vault.db ../release/audio-vault.db
cd ..
mkdir release/static-assets/
mkdir release/views/
#mkdir release/tools/
#mkdir release/vault/
cp ./audio-samples/segment*.wav ./release/
cp -R ./web-service/static-assets/* ./release/static-assets/
cp -R ./web-service/views/* ./release/views/
#cp -R ./web-service/vault/* ./release/vault/
#cp -R ./web-service/tools/* ./release/tools/
zip -9 -r audio-vault.zip release
mv audio-vault.zip ./release/
