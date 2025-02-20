.ONESHELL:

db:
	rm ./database/audio-vault.*
	cat ./database/schema.sql | sqlite3 ./database/audio-vault.db

test:
	cd web-service
	go build -ldflags="-w -s" -o audio-vault *.go
	cd ..
	./web-service/audio-vault

build:GOOS=windows
build:GOARCH=amd64
build:
	rm -r release/*
	cd web-service
	go build -ldflags="-w -s" -o audio-vault.exe *.go
	cd ..
	mkdir release/static-assets/
	mkdir release/views/
	cp ./web-service/audio-vault.exe ./release/audio-vault.exe
	cp -R ./web-service/static-assets/* ./release/static-assets/
	cp -R ./web-service/views/* ./release/views/
	zip -9 -r audio-vault.zip release
	mv audio-vault.zip ./release/