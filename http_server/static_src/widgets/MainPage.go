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

	document := dom.NewDocumentFromJsValue(gt.Get("document"))

	etc := elementtreeconstructor.NewElementTreeConstructor(
		document,
	)

	jrpc2_tester := NewJRPC2Tester(etc)
	arpc_tester := NewARPCTester(etc)

	new_html := etc.CreateElement("html").
		SetStyle("font-size", "10px").
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
					etc.CreateTextNode(" "),
					arpc_tester.Element,
				),
		)

	etc.ReplaceChildren(
		[]dom.ToNodeConvertable{
			document.Implementation().CreateDocumentType("html", "", ""),
			new_html,
		},
	)

	return
}
