package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/teeworlds-go/go-teeworlds-protocol/messages7"
	"github.com/teeworlds-go/go-teeworlds-protocol/protocol7"
)

func OnServerMessage(conn *Connection, twconn *protocol7.Connection, data []byte) []byte {
	packet := &protocol7.Packet{}
	err := packet.Unpack(data)
	if err != nil {
		panic(err)
	}

	// count sequence numbers so we can spoof the client
	// and send a disconnect message on behalf of the client
	twconn.OnPacket(packet)

	for i, msg := range packet.Messages {
		switch msg := msg.(type) {
		case *messages7.SvChat:
			// inspect server->client traffic
			Vlogf(0, "%s -> capitalism.\n", msg.Message)

			// change server->client traffic
			msg.Message = "capitalism."
			packet.Messages[i] = msg
			data = packet.Pack(twconn)
		case *messages7.CtrlToken:
			Vlogf(1, "got token=%x registering sigint handler ...\n", msg.Token)

			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, syscall.SIGINT)
			go func() {
				<-c
				Vlogf(0, "Got ctrl+c sending disconnect token=%x ack=%d ...\n", msg.Token, twconn.Ack)

				packet := protocol7.Packet{}
				packet.Messages = append(packet.Messages, &messages7.CtrlClose{})
				packet.Header.Token = msg.Token
				packet.Header.Ack = twconn.Ack
				disconnectPacked := packet.Pack(twconn)
				_, err = conn.ServerConn.Write(disconnectPacked)

				time.Sleep(10_000_000)

				Vlogf(0, "disconnected. token=%x ack=%d\n", msg.Token, twconn.Ack)
				Vlogf(0, "disconnect: %x\n", disconnectPacked)

				os.Exit(0)
			}()

		default:
		}
	}
	return data
}
