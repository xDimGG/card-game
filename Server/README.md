### Protocol of chat.go
The protocol is completely in plaintext
A WebSocket connection is opened after the client has connected to a lobby and received a JWT key

1. Client connects to WS on /chat
2. Client sends connection key
3. Server decodes JWT key and puts client in appropriate lobby
4. Server sends OK
5. Client may send messages at any time which will be broadcast to all lobby client in the format user_id:message
