package main

import (
	"encoding/json"
	"fmt"
	http_tracer "github.com/kacerr/http-tracer/http-tracer"
	"log"
	"os"
)

func main() {
	// time.Now()
	target := os.Args[1]

	// Create trace struct.
	t := http_tracer.New()
	target = "http://www.nhl.com"
	err := t.Get(target)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.MarshalIndent(t.Debug, "", "    ")
	fmt.Println(string(data))
}
