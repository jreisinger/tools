package tlsver

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type TLSversion uint16

func (tlsVersion TLSversion) String() string {
	switch tlsVersion {
	case 0:
		return ""
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return fmt.Sprintf("unknown %d", tlsVersion)
	}
}

type Getter struct {
	TCPaddr  string        // e.g. 1.1.1.1:443
	Timeout  time.Duration // TLS connection timeout
	Insecure bool          // don't verify the server's certificate
	TLSversion
	Err error
}

type option func(*Getter)

func WithTimeout(timeout time.Duration) option {
	return func(g *Getter) {
		g.Timeout = timeout
	}
}

func WithInsecure(insecure bool) option {
	return func(g *Getter) {
		g.Insecure = insecure
	}
}

func NewGetter(host, port string, opts ...option) *Getter {
	g := &Getter{
		TCPaddr:  net.JoinHostPort(host, port),
		Timeout:  10 * time.Second,
		Insecure: false,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *Getter) Get() {
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: g.Timeout},
		"tcp",
		g.TCPaddr,
		&tls.Config{InsecureSkipVerify: g.Insecure},
	)
	if err != nil {
		g.Err = err
		return
	}
	defer conn.Close()
	g.TLSversion = TLSversion(conn.ConnectionState().Version)
}
