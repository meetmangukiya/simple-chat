# simple-chat

`simple-chat` is a simple group chat implementation in go.

## Build

```bash
make
```

## Running

### Run server

```bash
./build/server -port 8080
```

### Run client

```bash
./build/client -host localhost -port 8080 -room golang -username gopher
```

## Implementation details

1. Uses websocket for realtime messages
2. In the backend, we maintain a mapping of room and all active connections to the room.
3. Whenever a message is received, the message is broadcasted to all the active room conncections using the mapping mentioned above.
4. Currently there is only one operation `sendMessage` that can be used by a user to send message. Structure of a generic websocket message is `{"op":"operation",params:["json", "params"]}`.

## Possible improvements

1. Implement some form of user authentication.
2. Implement some form of authorizations for access to rooms.
3. Moderation features like ability to kick users out.
