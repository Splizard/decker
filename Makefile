all:
	go build -o ./decker ./src
	
install:
	cp ./decker /usr/bin/decker
