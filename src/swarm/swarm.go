package swarm

import (
	"fmt"
	"gotor/peer"
	"net"
	"os"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

var conns []net.Conn

func Swarm() {
	conns = make([]net.Conn, 0, 16)

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			panic(err)
		}
	}(l)

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		conns = append(conns, conn)
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func bootstrap(peer peer.Peer) {
	//conn, err := net.Dial("tcp")
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)

	// Read the incoming connection into the buffer.
	for {
		nRead, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
		}

		for _, c := range conns {
			if c.RemoteAddr() == conn.RemoteAddr() {
				continue
			}
			c.Write(buf[:nRead])
		}

	}

	// Close the connection when you're done with it.
	conn.Close()
}
