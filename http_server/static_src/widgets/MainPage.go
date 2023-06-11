package widgets

import (
	"log"
	"syscall/js"

	"github.com/AnimusPEXUS/gojstools/elementtreeconstructor"
	"github.com/AnimusPEXUS/gojstools/webapi/dom"
)

type MainPage struct {
	site *Site
}

func NewMainPage(site *Site) *MainPage {
	ret := new(MainPage)
	return ret
}

func (self *MainPage) RenderMainPage() {
	log.Println("called RenderMainPage()")

	g := js.Global()

	gt := g.Get("globalThis")

	document := gt.Get("document")

	window_document := dom.NewDocumentFromJsValue(document)

	etc := elementtreeconstructor.NewElementTreeConstructor(
		window_document,
	)

	jrpc2_tester := NewJRPC2Tester(etc)

	new_html := etc.CreateElement("html").
		AppendChildren(
			etc.CreateElement("head").
				AppendChildren(
					etc.CreateElement("title").
						AppendChildren(
							etc.CreateTextNode("Hello!"),
						),
				),
			etc.CreateElement("body").
				AppendChildren(
					jrpc2_tester.Element,
				),
		)

	etc.ReplaceChildren([]dom.ToNodeConvertable{new_html})

	return
}
