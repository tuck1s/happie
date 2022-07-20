# happie
"Ping", via HAProxy, an Internet Endpoint

## Command-line usage
```
HAProxy tester
Usage of ./happie:
    ./happie [FLAGS] proxy:port source:port dest:port
    ports can be names (e.g. smtp) or numbers (e.g. 25).

proxy:port      Proxy listening for your request, e.g. 127.0.0.1:5000.

source:port     Address on the proxy used for onward conection.
                Must be an address hosted by the proxy itself, otherwise the request will fail.
                Set to :0 to have the proxy choose an ephemeral port.

dest:port       The service the proxy should connect to. e.g. 64.233.167.27:smtp (Google mail server).

Enclose IPv6 addresses with [] - e.g. [2a00:1450:400c:c0a::1b]:smtp.
    source and dest must belong to the same address family (both IPv4 or both IPv6)
    proxy addr can be of a different address family.

FLAGS:
  -v1
        Use PROXY protocol v1 header
```

## Building
```
cd cmd/happie
go build
```

## Examples
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

## IPv6 to the proxy

You can use `happie` to connect to your proxy via IPv6, without affecting the actual PROXY header content.

```
./happie [2600:1f13:b75:ba00:458d:f6da:c50:2ce3]:5000 172.31.15.167:0 172.31.4.82:smtp
Sending header version 2
00000000  0d 0a 0d 0a 00 0d 0a 51  55 49 54 0a 21 11 00 0c  |.......QUIT.!...|
00000010  ac 1f 0f a7 ac 1f 04 52  00 00 00 19              |.......R....|

Reply: 220 ip-172-31-4-82.us-west-2.compute.internal ESMTP service ready
```

Go expects the address to be within `[]` brackets, otherwise you will see an error message _too many colons in address_.

## IPv6 onward connection

You can also instruct the proxy to use IPv6 source and dest addresses in the onward connection.  The generated headers are longer.

```
./happie [2600:1f13:b75:ba00:458d:f6da:c50:2ce3]:5000 [2600:1f13:b75:ba00:53a9:f8b3:e450:2199]:0 [2600:1f13:b75:ba00:d473:1027:772:4170]:smtp
Sending header version 2
00000000  0d 0a 0d 0a 00 0d 0a 51  55 49 54 0a 21 21 00 24  |.......QUIT.!!.$|
00000010  26 00 1f 13 0b 75 ba 00  53 a9 f8 b3 e4 50 21 99  |&....u..S....P!.|
00000020  26 00 1f 13 0b 75 ba 00  d4 73 10 27 07 72 41 70  |&....u...s.'.rAp|
00000030  00 00 00 19                                       |....|

Reply: 220 ip-172-31-4-82.us-west-2.compute.internal ESMTP service ready
```

Using PROXY header version 1
```
./happie -v1 [2600:1f13:b75:ba00:458d:f6da:c50:2ce3]:5000 [2600:1f13:b75:ba00:53a9:f8b3:e450:2199]:0 [2600:1f13:b75:ba00:d473:1027:772:4170]:smtp
Sending header version 1
00000000  50 52 4f 58 59 20 54 43  50 36 20 32 36 30 30 3a  |PROXY TCP6 2600:|
00000010  31 66 31 33 3a 62 37 35  3a 62 61 30 30 3a 35 33  |1f13:b75:ba00:53|
00000020  61 39 3a 66 38 62 33 3a  65 34 35 30 3a 32 31 39  |a9:f8b3:e450:219|
00000030  39 20 32 36 30 30 3a 31  66 31 33 3a 62 37 35 3a  |9 2600:1f13:b75:|
00000040  62 61 30 30 3a 64 34 37  33 3a 31 30 32 37 3a 37  |ba00:d473:1027:7|
00000050  37 32 3a 34 31 37 30 20  30 20 32 35 0d 0a        |72:4170 0 25..|

Reply: 220 ip-172-31-4-82.us-west-2.compute.internal ESMTP service ready
```

PROXY protocol states the address family **once only** in the header. This means the _source_ and _dest_ addresses must belong to the same address family. If you try to mix them, you will see an error:

 `Source and dest addr must be both IPv4, or both IPv6 - cannot be mixed`.
