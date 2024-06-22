// Implementation of a UDP proxy

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/teeworlds-go/teeworlds/network7"
	"github.com/teeworlds-go/teeworlds/protocol7"
)

// Information maintained for each client/server connection
type Connection struct {
	ClientAddr *net.UDPAddr // Address of the client
	ServerConn *net.UDPConn // UDP connection to server
}

// Generate a new connection by opening a UDP connection to the server
func NewConnection(srvAddr, cliAddr *net.UDPAddr) *Connection {
	conn := new(Connection)
	conn.ClientAddr = cliAddr
	srvudp, err := net.DialUDP("udp", nil, srvAddr)
	if checkreport(1, err) {
		return nil
	}
	conn.ServerConn = srvudp
	return conn
}

// Global state
// Connection used by clients as the proxy server
var ProxyConn *net.UDPConn

// Address of server
var ServerAddr *net.UDPAddr

// Mapping from client addresses (as host:port) to connection
var ClientDict map[string]*Connection = make(map[string]*Connection)

// Mutex used to serialize access to the dictionary
var dmutex *sync.Mutex = new(sync.Mutex)

func setup(hostport string, port int) bool {
	// Set up Proxy
	saddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if checkreport(1, err) {
		return false
	}
	pudp, err := net.ListenUDP("udp", saddr)
	if checkreport(1, err) {
		return false
	}
	ProxyConn = pudp
	Vlogf(2, "Proxy serving on port %d\n", port)

	// Get server address
	srvaddr, err := net.ResolveUDPAddr("udp", hostport)
	if checkreport(1, err) {
		return false
	}
	ServerAddr = srvaddr
	Vlogf(2, "Connected to server at %s\n", hostport)
	return true
}

func dlock() {
	dmutex.Lock()
}

func dunlock() {
	dmutex.Unlock()
}

// Go routine which manages connection from server to single client
func RunConnection(conn *Connection, twconn *protocol7.Connection) {
	var buffer [1500]byte
	for {
		// Read from server
		n, err := conn.ServerConn.Read(buffer[0:])
		if checkreport(1, err) {
			continue
		}

		data := OnServerMessage(conn, twconn, buffer[0:n])

		// Relay it to client
		_, err = ProxyConn.WriteToUDP(data, conn.ClientAddr)
		if checkreport(1, err) {
			continue
		}
		Vlogf(3, "Relayed '%x' from server to %s.\n",
			buffer[0:n], conn.ClientAddr.String())
	}
}

// Routine to handle inputs to Proxy port
func RunProxy() {
	var buffer [1500]byte
	for {
		n, cliaddr, err := ProxyConn.ReadFromUDP(buffer[0:])
		if checkreport(1, err) {
			continue
		}
		Vlogf(3, "Read '%x' from client %s\n",
			buffer[0:n], cliaddr.String())
		saddr := cliaddr.String()
		dlock()
		conn, found := ClientDict[saddr]
		if !found {
			conn = NewConnection(ServerAddr, cliaddr)
			if conn == nil {
				dunlock()
				continue
			}
			ClientDict[saddr] = conn
			dunlock()
			Vlogf(0, "Created new connection for client %s\n", saddr)

			twconn := &protocol7.Connection{
				ClientToken: [4]byte{0x01, 0x02, 0x03, 0x04},
				ServerToken: [4]byte{0xff, 0xff, 0xff, 0xff},
				Ack:         0,
				Players:     make([]protocol7.Player, network7.MaxClients),
			}

			// Fire up routine to manage new connection
			go RunConnection(conn, twconn)

		} else {
			Vlogf(5, "Found connection for client %s\n", saddr)
			dunlock()
		}
		// Relay to server
		_, err = conn.ServerConn.Write(buffer[0:n])
		if checkreport(1, err) {
			continue
		}
	}
}

var verbosity int = 6

// Log result if verbosity level high enough
func Vlogf(level int, format string, v ...interface{}) {
	if level <= verbosity {
		log.Printf(format, v...)
	}
}

// Handle errors
func checkreport(level int, err error) bool {
	if err == nil {
		return false
	}
	Vlogf(level, "Error: %s", err.Error())
	return true
}

func main() {
	var ihelp *bool = flag.Bool("h", false, "Show help information")
	var ipport *int = flag.Int("p", 8333, "Proxy port")
	var isport *int = flag.Int("P", 8303, "Server port")
	var ishost *string = flag.String("H", "localhost", "Server address")
	var iverb *int = flag.Int("v", 1, "Verbosity (0-6)")
	//	var idrop *float64 = flag.Float64("d", 0.0, "Packet drop rate")
	flag.Parse()
	verbosity = *iverb
	if *ihelp {
		flag.Usage()
		os.Exit(0)
	}
	if flag.NArg() > 0 {
		ok := true
		fields := strings.Split(flag.Arg(0), ":")
		ok = ok && len(fields) == 2
		if ok {
			*ishost = fields[0]
			n, err := fmt.Sscanf(fields[1], "%d", isport)
			ok = ok && n == 1 && err == nil
		}
		if !ok {
			flag.Usage()
			os.Exit(0)
		}
	}
	hostport := fmt.Sprintf("%s:%d", *ishost, *isport)
	Vlogf(3, "Proxy port = %d, Server address = %s\n",
		*ipport, hostport)
	if setup(hostport, *ipport) {
		RunProxy()
	}
	os.Exit(0)
}
