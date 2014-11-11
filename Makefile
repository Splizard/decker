.PHONY: windows

all:
	go build -o ./decker ./src
	
32:
	GOARCH=386 \
	go build -o ./decker ./src	
	
install:
	cp ./decker /usr/bin/decker
	cp ./misc/decker.desktop /usr/share/applications/
	cp ./misc/mime.xml /usr/share/mime/packages/
	update-mime-database /usr/share/mime/
	update-desktop-database
	
windows:
	GOOS=windows \
	GOARCH=386 \
	go build -o ./windows/decker.exe ./src
	
deb:
	#Create the folders.
	mkdir -p ./pkg/deb/DEBIAN
	cp ./misc/version.info ./pkg/deb/DEBIAN/control
	if file decker | grep "64-bit"; then \
		sed "s/ARCHITECTURE/amd64/" -i ./pkg/deb/DEBIAN/control; fi
	
	if file decker | grep "32-bit"; then \
		sed "s/ARCHITECTURE/i386/" -i ./pkg/deb/DEBIAN/control; fi
	
	echo "#!/bin/sh -e\nupdate-desktop-database" >> ./pkg/deb/DEBIAN/postinst
	chmod +x ./pkg/deb/DEBIAN/postinst
	echo "2.0" > ./pkg/deb/debian-binary
	
	mkdir -p ./pkg/deb/sysroot/usr/bin/
	mkdir -p ./pkg/deb/sysroot/usr/share/applications/
	mkdir -p ./pkg/deb/sysroot/usr/share/mime/packages/
	
	#Copy files.
	cp ./decker ./pkg/deb/sysroot/usr/bin/
	chmod +x ./pkg/deb/sysroot/usr/bin/decker
	cp ./misc/decker.desktop ./pkg/deb/sysroot/usr/share/applications/
	cp ./misc/mime.xml ./pkg/deb/sysroot/usr/share/mime/packages/
	
	sed "s/SIZE/$(shell stat -c %s ./pkg/deb)/" -i ./pkg/deb/DEBIAN/control
	
	#Permissions.
	find ./pkg/deb/ -type d -exec chmod 0755 {} \;
	find ./pkg/deb/ -type f -exec chmod go-w {} \;
	chown -R root:root ./pkg/deb/
	
	cd ./pkg/deb/sysroot/ && tar czf ../data.tar.gz *
	cd ./pkg/deb/DEBIAN/ && tar czf ../control.tar.gz *
	
	find ./pkg/deb/ -type d -exec chmod 0755 {} \;
	find ./pkg/deb/ -type f -exec chmod go-w {} \;
	chown -R root:root ./pkg/deb/
	
	#Clean up and build.
	rm -rf ./pkg/deb/sysroot
	rm -rf ./pkg/deb/DEBIAN
	cd ./pkg/deb/ && ar r decker-0.5.deb debian-binary control.tar.gz data.tar.gz
	rm -f ./pkg/deb/control.tar.gz
	rm -f ./pkg/deb/data.tar.gz
	rm -f ./pkg/deb/debian-binary
	
	
