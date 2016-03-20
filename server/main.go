package main

import "log"

func main() {
	// set log flags
	log.SetFlags(log.Flags() | log.Lshortfile)
	s := NewServer()
	log.Println("** StartServer **")
	//Load cookie jar secret from file
	s.Init("secret")
	s.Serve()
}
