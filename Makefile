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

segments:
	curl http://localhost:1969/store/ -H "Authorization: cf83e1357eefb8bdf1542850d66d800" -v -F key1=value1 -F "fileupload=@./audio-samples/98767978-0999994H-12345-1.wav" -vvv
	curl http://localhost:1969/store/ -H "Authorization: cf83e1357eefb8bdf1542850d66d800" -v -F key1=value1 -F "fileupload=@./audio-samples/98767978-0999994H-67890-2.wav" -vvv