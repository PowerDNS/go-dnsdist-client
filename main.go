package main

import (
	"fmt"
	"github.com/PowerDNS/go-dnsdist-client/dnsdist"
	"log"
	"os"
)

func main() {
	dc, err := dnsdist.Dial(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatalf("Failure dialing: %s", err)
	}
	resp, err := dc.Command(os.Args[3])
	if err != nil {
		log.Fatalf("Failure executing command: %s", err)
	}
	fmt.Println(resp)
}
