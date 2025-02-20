db:
	rm ./database/audio-vault.*
	cat ./database/schema.sql | sqlite3 ./database/audio-vault.db

test:
	go build -ldflags="-w -s" -o ./web-service/audio-vault ./web-service/*.go
	./web-service/audio-vault

build:export GOOS=windows
build:export GOARCH=amd64
build:
	rm -r release/*
	go build -ldflags="-w -s" -o ./release/audio-vault.exe ./web-service/*.go
	mkdir release/static-assets/
	mkdir release/views/
	cp -R ./web-service/static-assets/* ./release/static-assets/
	cp -R ./web-service/views/* ./release/views/
	zip -9 -r audio-vault.zip release
	mv audio-vault.zip ./release/