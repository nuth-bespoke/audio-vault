db:
	rm ./database/audio-vault.*
	cat ./database/schema.sql | sqlite3 ./database/audio-vault.db

build:
	go build -ldflags="-w -s" -o ./web-service/audit-vault ./web-service/*.go
	./web-service/audit-vault
