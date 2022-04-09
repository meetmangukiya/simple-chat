all: server client

server: always
	cd server; go build -o ../build/server server.go config.go ws.go

client: always
	cd client; go build -o ../build/client client.go

always:
	echo "always"