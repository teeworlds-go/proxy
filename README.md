# teeworlds 0.7 MITM proxy

Man in the middle teeworlds proxy. Repacking chat messages. For now only the teeworlds 0.7 protocol is supported.

![go proxy banner](https://raw.githubusercontent.com/ChillerDragon/cdn/master/teeworlds_go_proxy_banner.png)

## run the proxy

```
go build
./proxy -H 127.0.0.1 -P 8303 -p 8333
```

Start a teeworlds 0.7 server on port 8303 on your machine.
Connect to 8333 with your 0.7 client and you will be proxied to the 8303 server.

## what it does

It uses a [teeworlds protocol implementation written in pure go](https://github.com/teeworlds-go/go-teeworlds-protocol) to unpack, change and then repack the traffic.
The current example changes all chat messages to "capitalism.". So a client connecting to the proxy will only see "capitalism." in the chat.

Using this command you can spin up a proxy server on localhost:8303 which connects to a ddnet CHINA server.
```
./proxy -H 101.43.114.27 -P 8327-p 8303
```

Once connected with a 0.7 client the chat will be full of capitalism messages.
```
DDNet7 "connect tw-0.7+udp://127.0.0.1:8303"
```

![mitm proxy capitalism chat](https://raw.githubusercontent.com/ChillerDragon/cdn/master/teeworlds_07_mitm_proxy_capitalism.png)

