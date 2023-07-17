package widgets

import (
	"errors"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/AnimusPEXUS/gojsonrpc2"
	"github.com/AnimusPEXUS/gojstools/elementtreeconstructor"
	"github.com/AnimusPEXUS/gojstools/std/array"
	"github.com/AnimusPEXUS/gojstools/webapi/dom"
	"github.com/AnimusPEXUS/gojstools/webapi/events"
	"github.com/AnimusPEXUS/gojstools/webapi/ws"
)

type JRPC2Tester struct {
	etc *elementtreeconstructor.ElementTreeConstructor

	ws   *ws.WS
	node *gojsonrpc2.JSONRPC2Node

	button_close *elementtreeconstructor.ElementMutator

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
	button_close := self.etc.CreateElement("button")
	button_ping := self.etc.CreateElement("button")

	self.button_close = button_close

	button_connect.AppendChildren(self.etc.CreateTextNode("Connect"))
	button_connect.Set(
		"onclick",
		js.FuncOf(
			func(this js.Value, args []js.Value) any {
				var err error

				self.cleanupWSAndNode()

				self.node = gojsonrpc2.NewJSONRPC2Node()
				self.node.PushMessageToOutsideCB =
					func(data []byte) (err error) {

						defer func() {
							if err != nil {
								self.AppendToConnectionLogText("err:", err.Error())
							}
						}()

						if self.ws == nil {
							return errors.New("self.ws not set")
						}

						arr, err := array.NewArrayFromByteSlice(data)
						if err != nil {
							return errors.New(
								"couldn't create new JS array: " + err.Error(),
							)
						}

						err = self.ws.Send(arr.JSValue)
						if self.ws == nil {
							return errors.New("self.ws.Send() err: " + err.Error())
						}

						return nil
					}
				self.node.OnRequestCB = self.onNodeRequest

				self.ws, err = ws.NewWS(
					&ws.WSOptions{
						URL: &([]string{"wss://localhost:8080/ws_jrpc2"}[0]),

						OnClose:   self.OnWSClose,
						OnError:   self.OnWSError,
						OnMessage: self.OnWSMessage,
						OnOpen:    self.OnWSOpen,
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
				return nil
			},
		),
	)

	button_close.AppendChildren(self.etc.CreateTextNode("Close"))
	button_close.Set("disabled", true)
	button_close.Set(
		"onclick",
		js.FuncOf(
			func(this js.Value, args []js.Value) any {
				self.cleanupWSAndNode()
				return nil
			},
		),
	)

	button_ping.AppendChildren(self.etc.CreateTextNode("Ping"))
	button_ping.Set(
		"onclick",
		js.FuncOf(
			func(this js.Value, args []js.Value) any {
				self.AppendToConnectionLogPreText(
					"self.ws ==", self.ws,
					"self.node ==", self.node,
				)
				if self.ws != nil && self.node != nil {
					err := self.node.SendNotification(
						&gojsonrpc2.Message{
							Request: gojsonrpc2.Request{
								Method: "ping",
							},
						},
					)

					self.AppendToConnectionLogPreText("ping error:", err)

				}
				return nil
			},
		),
	)

	self.Element.AppendChildren(
		self.etc.CreateElement("div").AppendChildren(
			self.etc.CreateTextNode("JSON-RPC 2 testing tool"),
		),
		self.etc.CreateElement("div").AppendChildren(
			button_connect,
			self.etc.CreateTextNode(" "),
			button_close,
			self.etc.CreateTextNode(" "),
			button_ping,
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

func (self *JRPC2Tester) AppendToConnectionLogText(
	v ...any,
) {
	x := self.etc.CreateTextNode(
		strings.TrimSpace(fmt.Sprintln(v...)),
	)

	self.AppendToConnectionLog(x)
}

func (self *JRPC2Tester) AppendToConnectionLogPreText(
	v ...any,
) {
	x := self.etc.CreateElement("pre").AppendChildren(
		self.etc.CreateTextNode(
			strings.TrimSpace(fmt.Sprintln(v...)),
		),
	)

	self.AppendToConnectionLog(x)
}

func (self *JRPC2Tester) cleanupWSAndNode() {
	if self.node != nil {
		self.node.Close()
		self.node = nil
	}

	if self.ws != nil {
		self.ws.Close()
		self.ws = nil
	}
}

func (self *JRPC2Tester) OnWSClose(e *events.CloseEvent) {
	self.AppendToConnectionLog(
		self.etc.CreateTextNode(
			fmt.Sprint("WS closed"),
		),
	)
	self.button_close.Set("disabled", true)
	self.cleanupWSAndNode()
}

func (self *JRPC2Tester) OnWSError(e *events.ErrorEvent) {
	self.AppendToConnectionLog(
		self.etc.CreateTextNode(
			fmt.Sprint("WS error"),
		),
	)
	self.cleanupWSAndNode()
}

func (self *JRPC2Tester) OnWSMessage(e *events.MessageEvent) {
	self.AppendToConnectionLog(
		self.etc.CreateTextNode(
			fmt.Sprint("WS message"),
		),
	)

	data, err := ws.GetByteSliceFromWSMessageEvent(e.JSValue)
	if err != nil {
		self.AppendToConnectionLogText(
			"err: problem while obtaining bytes from message:",
			err,
		)
		return
	}

	if self.node == nil {
		self.AppendToConnectionLogText(
			"err: self.node == nil",
		)
		return
	}

	proto_err, err := self.node.PushMessageFromOutside(data)
	if proto_err != nil || err != nil {
		self.AppendToConnectionLogText(
			"proto_err:", proto_err,
			"err:", err,
		)
	}
}

func (self *JRPC2Tester) OnWSOpen(e *events.Event) {
	self.AppendToConnectionLog(
		self.etc.CreateTextNode(
			fmt.Sprint("WS open"),
		),
	)

	self.button_close.Set("disabled", false)
}

func (self *JRPC2Tester) onNodeRequest(msg *gojsonrpc2.Message) (error, error) {
	if msg.IsRequestOrNotification() {
		what := ""
		if msg.IsNotification() {
			what = "Notification"
		} else {
			what = "Call"
		}
		self.AppendToConnectionLogText(fmt.Sprintf("got <%s>: %s(%#v)", what, msg.Method, msg.Params))
	}
	return nil, nil
}
