.ONESHELL:

db:
	rm ./database/audio-vault.*
	cat ./database/schema.sql | sqlite3 ./database/audio-vault.db


test: db
	cd web-service
	rm audio-vault.db*
	cp ../database/audio-vault.db audio-vault.db
	go build -ldflags="-w -s -X main.GIT_COMMIT_HASH=`git rev-parse HEAD`" -o audio-vault *.go
	./audio-vault


segments:
	curl http://localhost:1969/store/ -H "Authorization: cf83e1357eefb8bdf1542850d66d800" -v -F DocumentID=9999999999 -F MRN=0999994H -F CreatedBy=Paulx030 -F MachineName=SignalZero -F SegmentCount=2 -F SegmentOrder=1 -F "fileupload=@./audio-samples/9999999999-2-0999994H-12345-1.wav" -vvv
	sleep 3
	curl http://localhost:1969/store/ -H "Authorization: cf83e1357eefb8bdf1542850d66d800" -v -F DocumentID=9999999999 -F MRN=0999994H -F CreatedBy=Paulx030 -F MachineName=SignalZero -F SegmentCount=2 -F SegmentOrder=2 -F "fileupload=@./audio-samples/9999999999-2-0999994H-67890-2.wav" -vvv

poly:
	curl http://localhost:1969/store/ -H "Authorization: cf83e1357eefb8bdf1542850d66d800" -v -F DocumentID=987690000 -F MRN=0999994H -F CreatedBy=Paulx030 -F MachineName=SignalZero -F SegmentCount=3 -F SegmentOrder=1 -F "fileupload=@./audio-samples/segment-1.wav" -vvv
	# sleep 3
	# curl http://localhost:1969/store/ -H "Authorization: cf83e1357eefb8bdf1542850d66d800" -v -F DocumentID=987690000 -F MRN=0999994H -F CreatedBy=Paulx030 -F MachineName=SignalZero -F SegmentCount=3 -F SegmentOrder=2 -F "fileupload=@./audio-samples/segment-2.wav" -vvv
	# sleep 3
	# curl http://localhost:1969/store/ -H "Authorization: cf83e1357eefb8bdf1542850d66d800" -v -F DocumentID=987690000 -F MRN=0999994H -F CreatedBy=Paulx030 -F MachineName=SignalZero -F SegmentCount=3 -F SegmentOrder=3 -F "fileupload=@./audio-samples/segment-3.wav" -vvv