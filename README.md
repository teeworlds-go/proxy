# teeworlds proxy

WORK IN PROGRESS

## run the proxy

```
go build
./proxy -H 127.0.0.1 -P 8303 -p 8333
```

Start a teeworlds 0.7 server on port 8303 on your machine.
Connect to 8333 with your 0.7 client and you will be proxied to the 8303 server.

