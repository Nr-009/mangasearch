.PHONY: build start clean rebuild status search

build:
	go build -o mangasearch .

start: build
	./mangasearch start

clean:
	rm -f mangasearch

rebuild:
	./mangasearch rebuild-index

status:
	./mangasearch status

search:
	./mangasearch search "$(q)"