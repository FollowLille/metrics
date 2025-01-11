package main

import "os"

func main() {
	os.Exit(0) // want "can't use os.Exit in main function"
}
