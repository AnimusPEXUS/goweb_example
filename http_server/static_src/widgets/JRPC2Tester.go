package widgets

import (
	"syscall/js"

	"github.com/AnimusPEXUS/gojsonrpc2"
	"github.com/AnimusPEXUS/gojstools/elementtreeconstructor"
)

type JRPC2Tester struct {
	etc *elementtreeconstructor.ElementTreeConstructor

	node *gojsonrpc2.JSONRPC2Node

	connection_log *elementtreeconstructor.ElementMutator

	Element *elementtreeconstructor.ElementMutator
}

func NewJRPC2Tester(
	etc *elementtreeconstructor.ElementTreeConstructor,
) *JRPC2Tester {
	self := new(JRPC2Tester)

	self.etc = etc

	self.Element = self.etc.CreateElement("div")
	self.Element.SetStyle("border", "1px solid black")
	self.Element.SetStyle("padding", "5px")

	button_connect := self.etc.CreateElement("button")
	button_connect.AppendChildren(self.etc.CreateTextNode("Connect"))
	button_connect.Set(
		"onclick",
		js.FuncOf(
			func(this js.Value, args []js.Value) any {
				self.connection_log.AppendChildren(
					self.etc.CreateElement("div").AppendChildren(
						self.etc.CreateTextNode("connect clicked"),
					),
				)
				return nil
			},
		),
	)

	button_close := self.etc.CreateElement("button")
	button_close.AppendChildren(self.etc.CreateTextNode("Close"))

	self.Element.AppendChildren(
		self.etc.CreateElement("div").AppendChildren(
			self.etc.CreateTextNode("JSON-RPC 2 testing tool"),
		),
		self.etc.CreateElement("div").AppendChildren(
			button_connect,
			self.etc.CreateTextNode(" "),
			button_close,
		),
		self.etc.CreateElement("div").
			AppendChildren(
				self.etc.CreateTextNode("Connection Log"),
				self.etc.CreateElement("div").
					SetStyle("border", "1px solid black").
					SetStyle("max-height", "400px").
					SetStyle("overflow", "scroll").
					AssignSelf(&self.connection_log),
			),
	)

	return self
}
