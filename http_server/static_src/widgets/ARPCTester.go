package widgets

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"syscall/js"
	"time"

	"github.com/AnimusPEXUS/goarpcsolution"
	"github.com/AnimusPEXUS/goinmemfile"
	"github.com/AnimusPEXUS/gojsonrpc2datastreammultiplexer"
	"github.com/AnimusPEXUS/gojstools/elementtreeconstructor"
	"github.com/AnimusPEXUS/gojstools/std/array"
	"github.com/AnimusPEXUS/gojstools/webapi/dom"
	"github.com/AnimusPEXUS/gojstools/webapi/events"
	"github.com/AnimusPEXUS/gojstools/webapi/ws"
)

type ARPCTester struct {
	etc *elementtreeconstructor.ElementTreeConstructor

	ws          *ws.WS
	arpc        *goarpcsolution.ARPCNode
	arpc_basic  *goarpcsolution.ARPCNodeCtlBasic
	multiplexer *gojsonrpc2datastreammultiplexer.JSONRPC2DataStreamMultiplexer

	button_close *elementtreeconstructor.ElementMutator

	connection_log *elementtreeconstructor.ElementMutator

	Element *elementtreeconstructor.ElementMutator
}

func NewARPCTester(
	etc *elementtreeconstructor.ElementTreeConstructor,
) *ARPCTester {
	self := new(ARPCTester)

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

				self.cleanup()

				self.multiplexer =
					gojsonrpc2datastreammultiplexer.
						NewJSONRPC2DataStreamMultiplexer()

				self.multiplexer.PushMessageToOutsideCB =
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

				// self.multiplexer.OnRequestToProvideWriteSeekerCB = func(
				// 	size int64,
				// 	provide_data_destination func(io.WriteSeeker) error,
				// ) error {

				// 	buff := make([]byte, size)
				// 	imf := goinmemfile.NewInMemFileFromBytes(buff, 0, false)
				// 	provide_data_destination(imf)
				// 	return nil
				// }

				self.multiplexer.OnIncommingDataTransferComplete = func(wrs io.WriteSeeker) {
					imf, ok := wrs.(*goinmemfile.InMemFile)
					if !ok {
						panic("not InMemFile")
					}
					x := imf.Buffer

					// TODO: do something on error
					self.arpc.PushMessageFromOutside(x)
				}

				arpc_basic := goarpcsolution.NewARPCNodeCtlBasic()
				self.arpc_basic = arpc_basic

				arpc_basic.OnCallCB = self.onARPCCall

				self.arpc = goarpcsolution.NewARPCNode(arpc_basic)
				self.arpc.PushMessageToOutsideCB =
					func(data []byte) (err error) {

						timedout, closed, _, proto_err, err :=
							self.multiplexer.ChannelData(data)
						if err != nil {
							return err
						}

						if proto_err != nil {
							return proto_err
						}

						if timedout {
							return errors.New("timedout")
						}

						if closed {
							return errors.New("closed")
						}

						return nil
					}

				self.ws, err = ws.NewWS(
					&ws.WSOptions{
						URL: &([]string{"wss://localhost:8080/ws_arpc"}[0]),

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
				self.cleanup()
				return nil
			},
		),
	)

	button_ping.AppendChildren(self.etc.CreateTextNode("File List"))
	button_ping.Set(
		"onclick",
		js.FuncOf(
			func(this js.Value, args []js.Value) any {
				self.AppendToConnectionLogPreText(
					"self.ws ==", self.ws,
					"self.arpc ==", self.arpc,
					"self.arpc_basic ==", self.arpc_basic,
				)
				if self.ws != nil &&
					self.arpc != nil &&
					self.arpc_basic != nil {
					// call, err := goarpcsolution.NewARPCCall()
					// if err != nil {
					// 	panic("programming error: " + err.Error())
					// }

					timedout, closed, call, rh :=
						goarpcsolution.NewChannelledARPCNodeCtlBasicRespHandler()

					_, err := self.arpc_basic.Call(
						"getfilelist",
						[]*goarpcsolution.ARPCCallArg{},
						false,
						rh,
						time.Minute*10,
					)
					if err != nil {
						self.AppendToConnectionLogPreText("filelist timeout")
						return nil
					}

					select {
					case <-timedout:
						self.AppendToConnectionLogPreText("filelist timeout")
					case <-closed:
						self.AppendToConnectionLogPreText("filelist connection closed")
					case <-call:
						self.AppendToConnectionLogPreText("filelist ok")
					}

				}
				return nil
			},
		),
	)

	self.Element.AppendChildren(
		self.etc.CreateElement("div").AppendChildren(
			self.etc.CreateTextNode("ARPC testing tool"),
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

func (self *ARPCTester) AppendToConnectionLog(
	e dom.ToNodeConvertable,
) {
	self.connection_log.AppendChildren(
		self.etc.CreateElement("div").AppendChildren(
			e,
		),
	)
	self.connection_log.Set("scrollTop", self.connection_log.Get("scrollTopMax"))
}

func (self *ARPCTester) AppendToConnectionLogText(
	v ...any,
) {
	x := self.etc.CreateTextNode(
		strings.TrimSpace(fmt.Sprintln(v...)),
	)

	self.AppendToConnectionLog(x)
}

func (self *ARPCTester) AppendToConnectionLogPreText(
	v ...any,
) {
	x := self.etc.CreateElement("pre").AppendChildren(
		self.etc.CreateTextNode(
			strings.TrimSpace(fmt.Sprintln(v...)),
		),
	)

	self.AppendToConnectionLog(x)
}

func (self *ARPCTester) cleanup() {

	if self.arpc != nil {
		self.arpc.Close()
		self.arpc = nil
	}

	if self.arpc_basic != nil {
		self.arpc_basic = nil
	}

	if self.multiplexer != nil {
		self.multiplexer.Close()
		self.multiplexer = nil
	}

	if self.ws != nil {
		self.ws.Close()
		self.ws = nil
	}
}

func (self *ARPCTester) OnWSClose(e *events.CloseEvent) {
	self.AppendToConnectionLog(
		self.etc.CreateTextNode(
			fmt.Sprint("WS closed"),
		),
	)
	self.button_close.Set("disabled", true)
	self.cleanup()
}

func (self *ARPCTester) OnWSError(e *events.ErrorEvent) {
	self.AppendToConnectionLog(
		self.etc.CreateTextNode(
			fmt.Sprint("WS error"),
		),
	)
	self.cleanup()
}

func (self *ARPCTester) OnWSMessage(e *events.MessageEvent) {
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

	if self.multiplexer == nil {
		self.AppendToConnectionLogText(
			"err: self.multiplexer == nil",
		)
		return
	}

	proto_err, err := self.multiplexer.PushMessageFromOutside(data)
	if proto_err != nil || err != nil {
		self.AppendToConnectionLogText(
			"proto_err:", proto_err,
			"err:", err,
		)
	}
}

func (self *ARPCTester) OnWSOpen(e *events.Event) {
	self.AppendToConnectionLog(
		self.etc.CreateTextNode(
			fmt.Sprint("WS open"),
		),
	)

	self.button_close.Set("disabled", false)
}

func (self *ARPCTester) onARPCCall(call *goarpcsolution.ARPCCall) (error, error) {
	log.Println("onARPCCall")
	return nil, nil
}
