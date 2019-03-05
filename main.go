package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/PowerDNS/go-dnsdist-client/dnsdist"
)

func main() {
	var host = flag.String("host", "127.0.0.1:5199", "host:port to connect to")
	var key = flag.String("key", "", "shared secret for the console")
	flag.Parse()
	dc, err := dnsdist.Dial(*host, *key)

	if err != nil {
		log.Fatalf("Failure dialing: %s", err)
	}
	for _, cmd := range flag.Args() {
		resp, err := dc.Command(cmd)
		if err != nil {
			log.Fatalf("Failure executing command: %s", err)
		}
		fmt.Print(resp)
	}
}
