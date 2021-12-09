package tcp

import (
	"bufio"
	"math/rand"
	"net"
	"strconv"
	"testing"
)

func TestListenAndServe(t *testing.T) {
	closeChan := make(chan struct{})
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	go ListenAndServe(listener, NewEchoHandler(), closeChan)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		val := strconv.Itoa(rand.Int())
		_, err := conn.Write([]byte(val + "\n"))
		if err != nil {
			t.Fatal(err)
		}
		bufReader := bufio.NewReader(conn)
		res, err := bufReader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if res != val+"\n" {
			t.Fatalf("expected %s, got %s", val, res)
		}
	}
}
