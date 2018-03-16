package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"./config"
	"./server"
)

func main() {
	version := fmt.Sprintf("%d.%d.%d", config.VersionMajor, config.VersionMinor, config.VersionPatch)

	if len(os.Args) > 1 {
		if os.Args[1] == "-v" {
			fmt.Println("botsbox version:", version)
			return
		} else if os.Args[1] == "-h" {
			fmt.Println("Usage: botsbox [options]")
			fmt.Println("  options:")
			fmt.Println("    -h:    Show help")
			fmt.Println("    -v:    Print version")
			return
		}
	}

	flag.Parse()

	ok := server.Start()
	if ok {
		log.Printf("### botsbox %s started \n", version)
	}
}
