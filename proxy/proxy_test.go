package proxy

import (
	"net"
	"strconv"
	"testing"
)

func TestWrapNoProxyUsesDirectDialer(t *testing.T) {
	t.Setenv("ALL_PROXY", "socks5://127.0.0.1:1")
	t.Setenv("all_proxy", "")
	t.Setenv("NO_PROXY", "localhost")
	t.Setenv("no_proxy", "")

	var dialer Dialer
	if !Wrap(func(value Dialer) { dialer = value }) {
		t.Fatal("expected proxy dialer")
	}
	if dialer == nil {
		t.Fatal("expected non-nil proxy dialer")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	accepted := make(chan struct{})
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = connection.Close()
		}
		close(accepted)
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	connection, err := dialer.Dial("tcp", net.JoinHostPort("localhost", strconv.Itoa(port)))
	if err != nil {
		t.Fatalf("dial NO_PROXY address directly: %v", err)
	}
	_ = connection.Close()
	<-accepted
}
