package exporter

import (
	"flag"
	"log"
	"strings"
	"regexp"
)

var (
	ready   = make(chan interface{}, 1)
	targets = make([]*target, 0)
	port    int
	
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var (
		skipDb, skipProd, skipNet bool
		nodes                     string
	)
	flag.StringVar(&nodes, "u", "http://localhost:8888", "nodes to monitor, comma seperated list of http urls with optional {labels} to override hostname")
	flag.BoolVar(&skipProd, "no-producer", false, "do not attempt to use the producer API")
	flag.BoolVar(&skipDb, "no-db", false, "do not attempt to use the db API")
	flag.BoolVar(&skipNet, "no-net", false, "do not attempt to use the net API")
	flag.IntVar(&port, "p", 13856, "port to listen on")
	flag.Parse()

	regexer, _ := regexp.Compile(`^\{.+\}`)

	urls := strings.Split(nodes, ",")
	for _, targetUrl := range urls {

		log.Printf("Processing target %s", targetUrl)
		
		label := regexer.FindString(targetUrl)

		if len(label) > 0 {
			targetUrl = strings.Replace(targetUrl, label, "", 1)
		}

		t, e := newTarget(targetUrl, label)
		if e != nil {
			log.Printf("could not connect to %s: %v. Host WILL NOT be monitored", targetUrl, e)
			continue
		}
		switch false {
		case skipProd:
			t.check("producer")
			fallthrough
		case skipDb:
			t.check("db")
			fallthrough
		case skipNet:
			t.check("net")
		}
		targets = append(targets, t)
	}
	if len(targets) == 0 {
		flag.PrintDefaults()
		log.Fatal("could not connect to any targets.")
	}
	close(ready)
}
