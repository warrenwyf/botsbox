package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const versionMajor int = 0
const versionMinor int = 0
const versionPatch int = 1

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-v" {
			fmt.Println(fmt.Sprintf("botsbox version: %d.%d.%d", versionMajor, versionMinor, versionPatch))
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

	log.Printf("### botsbox %d.%d.%d started", versionMajor, versionMinor, versionPatch)
}
