package widgets

import (
	"log"
	"syscall/js"

	"github.com/AnimusPEXUS/gojstools/elementtreeconstructor"
	"github.com/AnimusPEXUS/gojstools/webapi/dom"
	"github.com/AnimusPEXUS/gojstools/widgetcollection"
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

	login_form := widgetcollection.NewLoginPasswordForm00(
		etc,
		"", "",
		true,
		func() { log.Println("something edited") },
		func() { log.Println("login edited") },
		func() { log.Println("password edited") },
		func() { log.Println("button clicked") },
	)

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
					login_form.Element,
				),
		)

	etc.ReplaceChildren([]dom.ToNodeConvertable{new_html})

	return
}
