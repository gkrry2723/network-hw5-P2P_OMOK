# network-hw5-P2P_OMOK
P2P omok game application using TCP and UDP socket

### 1. Rendezvous Server
- TCP Server
- accept connections from clients that wish to play (게임 대기방 - 상대방 매칭을 해줌)
- When there is 0 or 1 client connected, it waits.
- As soon As there are 2 clients, it notifies both clients about the info of each other(IP, UDP port number, nickname ..)
- disconnect them


### 2. Client/Peer
- UDP peer for playing omok game(using UDP socket)
