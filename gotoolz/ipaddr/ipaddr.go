package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gotoolz/ipaddr/info"
)

func main() {
	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	flag.Parse()

	ip := net.ParseIP(flag.Arg(0))
	if ip == nil {
		log.Fatalf("not a valid IP address: %q", flag.Arg(0))
	}

	info, err := info.Get(ip)
	if err != nil {
		log.Print(err)
	}

	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
