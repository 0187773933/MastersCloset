#!/bin/bash

sudo chown -R morphs:morphs /home/morphs/SAVE_FILES

HASH_FILE="/home/morphs/git.hash"
REMOTE_HASH=$(git ls-remote https://github.com/0187773933/MastersCloset.git HEAD | awk '{print $1}')

if [ -f "$HASH_FILE" ]; then
	STORED_HASH=$(sudo cat "$HASH_FILE")
else
	STORED_HASH=""
fi

if [ "$REMOTE_HASH" == "$STORED_HASH" ]; then
	echo "No New Updates Available"
	cd /home/morphs/MastersCloset
	LOG_LEVEL=debug exec /home/morphs/MastersCloset/server "$@"
else
	echo "New updates available. Updating and Rebuilding Go Module"
	echo "$REMOTE_HASH" | sudo tee "$HASH_FILE"
	cd /home/morphs
	sudo rm -rf /home/morphs/MastersCloset
	git clone "https://github.com/0187773933/MastersCloset.git"
	sudo chown -R morphs:morphs /home/morphs/MastersCloset
	cd /home/morphs/MastersCloset
	/usr/local/go/bin/go mod tidy
	GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o /home/morphs/MastersCloset/server
	LOG_LEVEL=debug exec /home/morphs/MastersCloset/server "$@"
fi