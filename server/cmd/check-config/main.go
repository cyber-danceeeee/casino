// check-config validates startup configuration without connecting to business services.
package main

import (
	"flag"
	"fmt"
	"log"

	"casino/internal/config"
)

func main() {
	filename := flag.String("c", "configs/casino.yml", "configuration file (also anchors relative paths for Nacos)")
	flag.Parse()
	if err := config.Init(*filename); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Configuration validated: source=%s, HTTP port=%d.\n", config.Source(), config.Get().HTTP.Port)
}
