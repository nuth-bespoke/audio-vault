.ONESHELL:

db:
	rm ./database/audio-vault.*
	cat ./database/schema.sql | sqlite3 ./database/audio-vault.db
	cp ./database/audio-vault.db ./web-service/audio-vault.db

test:
	cd web-service
	go build -ldflags="-w -s -X main.GIT_COMMIT_HASH=`git rev-parse HEAD`" -o audio-vault *.go
	cd ..
	./web-service/audio-vault
