.PHONY: windows

all:
	go build -o ./decker ./src
	
install:
	cp ./decker /usr/bin/decker
	
windows:
	GOOS=windows \
	GOARCH=386 \
	go build -o ./decker.exe ./src
