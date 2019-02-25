package main

import (
	"fmt"
	"github.com/powerdns/go-dnsdist-client/dnsdist"
	"log"
	"os"
)

func main() {
	dc, err := dnsdist.Dial(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatalf("Failure dialing: %s", err)
	}
	fmt.Println(dc)
	resp, err := dc.Command(os.Args[3])
	if err != nil {
		log.Fatal("Failure executing command: %s", err)
	}
	fmt.Println(resp)
}
