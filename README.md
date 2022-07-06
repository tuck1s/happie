# happie
"Ping" a TCP endpoint via HAProxy instance

## Usage

```
./happie
HAProxy tester
Usage of ./happie:
    ./happie [FLAGS] proxy:port source:port dest:port
    ports can be names (e.g. smtp) or numbers (e.g. 25).

proxy:port      Proxy listening for your request, e.g. 127.0.0.1:5000.

source:port     Address on the proxy used for onward conection.
                Must be an address hosted by the proxy itself, otherwise the request will fail.
                Set to :0 to have the proxy choose an ephemeral port.

dest:port       The service the proxy should connect to. e.g. 64.233.167.27:smtp (Google mail server).

FLAGS:
  -v1
        Use PROXY protocol v1 header
```

## Command-line utlity

Build the tool:
```
cd cmd/happie
go build
```

Ping a Gmail server via a proxy using SMTP - using PROXY header version 2
```
./happie 35.90.110.253:5000 172.31.15.167:0 64.233.167.27:smtp
Sending header version 2
00000000  0d 0a 0d 0a 00 0d 0a 51  55 49 54 0a 21 11 00 0c  |.......QUIT.!...|
00000010  ac 1f 0f a7 40 e9 a7 1b  00 00 00 19              |....@.......|

Reply: 220 mx.google.com ESMTP ca7-20020a056000088700b0021b96403933si45674690wrb.956 - gsmtp
```

Using PROXY header version 1
```
 ./happie -v1 35.90.110.253:5000 172.31.15.167:0 64.233.167.27:smtp
Sending header version 1
00000000  50 52 4f 58 59 20 54 43  50 34 20 31 37 32 2e 33  |PROXY TCP4 172.3|
00000010  31 2e 31 35 2e 31 36 37  20 36 34 2e 32 33 33 2e  |1.15.167 64.233.|
00000020  31 36 37 2e 32 37 20 30  20 32 35 0d 0a           |167.27 0 25..|

Reply: 220 mx.google.com ESMTP 205-20020a1c02d6000000b003a159e6b143si23380168wmc.221 - gsmtp
```