package main

import (
	"log"

	"github.com/AnimusPEXUS/goweb_example/http_server/static_src/widgets"
)

func main() {

	log.SetFlags(log.Lshortfile)

	log.Println("ExampleSite loading..")
	log.Println("  build time:", GOWEB_BUILD_TIME)
	log.Println("      commit:", GOWEB_BUILD_COMMIT)

	site := widgets.NewSite()

	site.ApplyToDocument()

	log.Println("  ExampleSite running")

	c := make(chan struct{})
	<-c
}
