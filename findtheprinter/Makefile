LDFLAGS="-s -w"

build: clean
	go build -ldflags $(LDFLAGS) -o build/findtheprinter findtheprinter.go

install:
	cp build/findtheprinter /usr/local/bin/
	chmod u+s /usr/local/bin/findtheprinter

clean:
	rm -rf ./build