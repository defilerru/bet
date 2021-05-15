package main

import (
	"flag"
	"log"
	"net/http"
)

type strings []string

func (i *strings) String() string {
	return "my string representation"
}

func (i *strings) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *strings) Contain(value string) bool {
	for n, _ := range *i {
		if (*i)[n] == value {
			return true
		}
	}
	return false
}

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var allowedOrigins strings

func main() {
	var err error
	flag.Var(&allowedOrigins, "allowed-origin", "may be specified multiple times")
	flag.Parse()
	log.SetFlags(0)

	db, err = NewMySQLDB("defiler@/defiler?parseTime=true&loc=Local")
	if err != nil {
		log.Fatalf("unable to connect db: %s", err)
	}
	activePredictions, err = db.LoadPredictions()
	if err != nil {
		log.Fatalf("unable to load predicionts: %s", err)
	}
	log.Printf("%d active predictions loaded", len(activePredictions))

	http.HandleFunc("/echo/", echo)
	fs := http.FileServer(http.Dir("html"))
	http.Handle("/", fs)
	log.Printf("Starting server at %s", *addr)

	log.Fatal(http.ListenAndServe(*addr, nil))
}