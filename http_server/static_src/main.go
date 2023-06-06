package main

import (
	"log"

	"github.com/AnimusPEXUS/goweb_example/http_server/page/widgets"
)

func main() {

	log.SetFlags(log.Lshortfile)

	log.Println("ExampleSite loading..")

	site := widgets.NewSite()

	site.ApplyToDocument()

	log.Println("  ExampleSite running")

	c := make(chan struct{})
	<-c
}
