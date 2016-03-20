package main

import (
	"encoding/base64"
	"fmt"

	"github.com/gorilla/securecookie"
)

func main() {
	fmt.Println(base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)))
}
