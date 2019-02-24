package main

import (
	"fmt"
	"github.com/powerdns/go-dnsdist-client/dnsdist"
	"log"
	"os"
)

func main() {
	dc, err := dnsdist.Connect(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dc)
	resp, err := dc.Command("return showVersion()")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}
