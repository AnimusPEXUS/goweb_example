package widgets

import (
	"fmt"
	"syscall/js"

	"github.com/AnimusPEXUS/gojsonrpc2"
	"github.com/AnimusPEXUS/gojstools/elementtreeconstructor"
	"github.com/AnimusPEXUS/gojstools/webapi/dom"
	"github.com/AnimusPEXUS/gojstools/webapi/events"
	"github.com/AnimusPEXUS/gojstools/webapi/ws"
)

type JRPC2Tester struct {
	etc *elementtreeconstructor.ElementTreeConstructor

	ws   *ws.WS
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
				var err error
				self.ws, err = ws.NewWS(
					&ws.WSOptions{
						URL: &([]string{"http://localhost:8080/ws_jrpc2"}[0]),

						OnClose: func(*events.CloseEvent) {
							self.AppendToConnectionLog(
								self.etc.CreateTextNode(
									fmt.Sprint(
										"WS closed",
										err,
									),
								),
							)
						},
						OnError: func(*events.ErrorEvent) {
							self.AppendToConnectionLog(
								self.etc.CreateTextNode(
									fmt.Sprint(
										"WS error",
										err,
									),
								),
							)
						},
						OnMessage: func(e *events.MessageEvent) {
							self.AppendToConnectionLog(
								self.etc.CreateTextNode(
									fmt.Sprint(
										"WS message",
										err,
									),
								),
							)
							val, err := e.GetData()
							self.AppendToConnectionLog(
								self.etc.CreateElement("pre").
									AppendChildren(
										self.etc.CreateTextNode(
											fmt.Sprint(
												"val:", val, "err:", err,
											),
										),
									),
							)
						},
						OnOpen: func(*events.Event) {
							self.AppendToConnectionLog(
								self.etc.CreateTextNode(
									fmt.Sprint(
										"WS open",
										err,
									),
								),
							)
						},
					},
				)
				if err != nil {
					self.AppendToConnectionLog(
						self.etc.CreateTextNode(
							fmt.Sprint(
								"WS connection error:",
								err,
							),
						),
					)
					return nil
				}
				self.node = gojsonrpc2.NewJSONRPC2Node()
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
					SetStyle("max-height", "200px").
					SetStyle("overflow", "scroll").
					AssignSelf(&self.connection_log),
			),
	)

	return self
}

func (self *JRPC2Tester) AppendToConnectionLog(
	e dom.ToNodeConvertable,
) {
	self.connection_log.AppendChildren(
		self.etc.CreateElement("div").AppendChildren(
			e,
		),
	)
	self.connection_log.Set("scrollTop", self.connection_log.Get("scrollTopMax"))
}
