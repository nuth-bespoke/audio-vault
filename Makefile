db:
	rm ./database/audio-vault.*
	cat ./database/schema.sql | sqlite3 ./database/audio-vault.db

build:
	go build -ldflags="-w -s" -o ./web-service/audio-vault ./web-service/*.go
	./web-service/audio-vault


deploy:export GOOS=windows
deploy:export GOARCH=amd64
deploy:
	go build -ldflags="-w -s" -o ./release/audio-vault.exe ./web-service/*.go