/*
Tlsver gets TLS version of a host.

Usage:

	$ tlsver go.dev perl.org
	$ subfinder -d go.dev --silent | tlsver -insecure -timeout 10s
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jreisinger/tools/tlsver"
)

var (
	concurrency = flag.Int("concurrency", 10, "maximum number of concurrent connections")
	insecure    = flag.Bool("insecure", false, "don't validate certificate")
	port        = flag.String("port", "443", "TCP port to connect")
	timeout     = flag.Duration("timeout", 5*time.Second, "TLS connection timeout")
)

func main() {
	flag.Parse()

	in := make(chan *tlsver.Getter)
	out := make(chan *tlsver.Getter)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		if len(flag.Args()) > 0 {
			for _, host := range flag.Args() {
				g := tlsver.NewGetter(
					host,
					*port,
					tlsver.WithInsecure(*insecure),
					tlsver.WithTimeout(*timeout),
				)
				in <- g
			}
		} else {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				host := s.Text()
				g := tlsver.NewGetter(
					host,
					*port,
					tlsver.WithInsecure(*insecure),
					tlsver.WithTimeout(*timeout),
				)
				in <- g
			}
		}
		close(in)
		wg.Done()
	}()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			for g := range in {
				g.Get()
				out <- g
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	for g := range out {
		if g.Err != nil {
			fmt.Fprintf(os.Stderr, "tlsver: %s: %v\n", g.TCPaddr, g.Err)
			continue
		}
		fmt.Printf("%s\t%s\n", g.TLSversion, g.TCPaddr)
	}
}
