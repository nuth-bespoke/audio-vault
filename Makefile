.ONESHELL:

db:
	rm ./database/audio-vault.*
	cat ./database/schema.sql | sqlite3 ./database/audio-vault.db

test:
	cd web-service
	go build -ldflags="-w -s" -o audio-vault *.go
	cd ..
	./web-service/audio-vault
