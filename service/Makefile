
all:	heapdump

heapdump:
	go mod tidy
	go build -o heapdump .

docker:
	docker build -t heapdump:latest .

clean:
	rm -f thumbnailer
	docker rmi -f heapdump:latest 2>/dev/null || /bin/true

test:
	go mod tidy || /bin/true
	go test -v ./...
	
bdd:
	godog run -t ~@wip

