package main

import (
	"fmt"
	"log"
	"time"

	"github.com/AnimusPEXUS/goweb_example/http_server/static_src/widgets"
)

func main() {

	log.SetFlags(log.Lshortfile)

	log.Println("ExampleSite loading..")

	t, err := time.Parse(time.RFC3339Nano, GOWEB_BUILD_TIME)
	if err != nil {
		panic(err)
	}

	how_long := time.Now().UTC().Sub(t)

	log.Println(
		fmt.Sprintf(
			"  build time: %s (%f days ago (%f hours))",
			t.Format(time.RFC3339Nano),
			how_long.Hours()/24,
			how_long.Hours(),
		),
	)
	log.Println("      commit:", GOWEB_BUILD_COMMIT)

	site := widgets.NewSite()

	site.ApplyToDocument()

	log.Println("  ExampleSite running")

	c := make(chan struct{})
	<-c
}
