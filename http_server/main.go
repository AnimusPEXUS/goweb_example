package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/AnimusPEXUS/goarpcsolution"
	"github.com/AnimusPEXUS/goinmemfile"
	"github.com/AnimusPEXUS/gojsonrpc2"
	"github.com/AnimusPEXUS/gojsonrpc2datastreammultiplexer"
	"github.com/julienschmidt/httprouter"
)

type Controller struct {
	Name string
}

func (self *Controller) HelloPage(
	rw http.ResponseWriter,
	rq *http.Request,
	params httprouter.Params,
) {
	var name_part string

	name_part = "<noname>"
	if self.Name != "" {
		name_part = self.Name
	}

	additional_message := ""

	switch rq.Method {
	default:
		additional_message =
			fmt.Sprintf("unsupported method: %s", rq.Method)
	case "GET":
		query := rq.URL.Query()

		names, ok := query["name"]
		if ok {
			name_part = names[0]
			self.Name = name_part
		}
	case "POST":
		rq.ParseForm()
		values := rq.PostForm

		names, ok := values["name"]
		if ok {
			name_part = names[0]
			self.Name = name_part
		}
	}

	if additional_message == "" {
		additional_message =
			fmt.Sprintf("(Used method is %s)", rq.Method)
	}

	rw.Write(
		[]byte(
			fmt.Sprintf(
				`<!DOCTYPE html>
<html>
 <head>
  <title>Hello %s</title>
 </head>
 <body>
 %[2]s
  <div>your name is %[1]s, yes?</div>
  <form action="/hello">
    <input type="text" name="name" />
	<input type="submit" /> (send using GET method)
  </form>
  <form action="/hello" method="POST">
    <input type="text" name="name" />
	<input type="submit" /> (send using POST method)
  </form>
 </body>
</html>
`,
				html.EscapeString(name_part),
				html.EscapeString(additional_message),
			),
		),
	)
}

func (self *Controller) WSHandlerJRPC2(c *websocket.Conn) {

	defer c.Close()

	codec := GenDirectWSCodec()

	node := gojsonrpc2.NewJSONRPC2Node()
	node.OnRequestCB = func(msg *gojsonrpc2.Message) (error, error) {
		if msg.IsNotification() {
			fmt.Println(
				"got notification:", msg.Method, "params:", msg.Params,
			)
		}
		return nil, nil
	}

	defer node.Close()

	var (
		stop_flag     = false
		rec_err       error
		send_err      error
		rec_proto_err error
	)

	wg := &sync.WaitGroup{}

	node.PushMessageToOutsideCB = func(data []byte) error {
		wg.Add(1)
		defer wg.Done()

		if stop_flag {
			return errors.New("stop_flag is true")
		}

		err := codec.Send(c, data)
		if err != nil {
			send_err = err
			return err
		}

		return nil
	}

	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			if stop_flag {
				return
			}

			var data []byte
			err := codec.Receive(c, &data)
			if err != nil {
				rec_err = err
				stop_flag = true
				return
			}

			proto_err, err := node.PushMessageFromOutside(data)
			if proto_err != nil || err != nil {
				rec_proto_err = proto_err
				rec_err = err
				stop_flag = true
				return
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			time.Sleep(time.Second)
			if stop_flag {
				return
			}
			node.SendNotification(
				&gojsonrpc2.Message{
					Request: gojsonrpc2.Request{
						Method: "time",
						Params: time.Now().UTC().Format(
							time.RFC3339Nano,
						),
					},
				},
			)
		}
	}()

	wg.Wait()

	fmt.Println("rec_err:", rec_err)
	fmt.Println("send_err:", send_err)
	fmt.Println("rec_proto_err:", rec_proto_err)

}

func (self *Controller) WSHandlerARPC(c *websocket.Conn) {

	defer c.Close()

	codec := GenDirectWSCodec()

	multiplexer :=
		gojsonrpc2datastreammultiplexer.NewJSONRPC2DataStreamMultiplexer()

	defer multiplexer.Close()

	arpc_basic := goarpcsolution.NewARPCNodeCtlBasic()

	arpc := goarpcsolution.NewARPCNode(arpc_basic)

	defer arpc.Close()

	var (
		stop_flag     = false
		rec_err       error
		send_err      error
		rec_proto_err error
	)

	multiplexer.OnIncommingDataTransferComplete = func(ws io.WriteSeeker) {

		if stop_flag {
			return
		}

		f, ok := ws.(*goinmemfile.InMemFile)
		if !ok {
			panic("programming error")
		}

		x := f.Buffer

		proto_err, err := arpc.PushMessageFromOutside(x)
		if proto_err != nil {
			rec_err = proto_err
			stop_flag = true
			return
		}

		if err != nil {
			rec_err = err
			stop_flag = true
			return
		}
	}

	arpc.PushMessageToOutsideCB = func(data []byte) error {
		timedout, closed, _, proto_err, err :=
			multiplexer.ChannelData(data)
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

	arpc_basic.OnCallCB = self.HandleARPCCall

	wg := &sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			if stop_flag {
				return
			}

			var data []byte
			err := codec.Receive(c, &data)
			if err != nil {
				rec_err = err
				stop_flag = true
				return
			}

			proto_err, err := multiplexer.PushMessageFromOutside(data)
			if proto_err != nil || err != nil {
				rec_proto_err = proto_err
				rec_err = err
				stop_flag = true
				return
			}
		}
	}()

	go func() {

	}()

	wg.Wait()

	fmt.Println("rec_err:", rec_err)
	fmt.Println("send_err:", send_err)
	fmt.Println("rec_proto_err:", rec_proto_err)

}

func GenDirectWSCodec() *websocket.Codec {
	codec := &websocket.Codec{
		Marshal: func(
			v interface{},
		) (
			data []byte,
			payloadType byte,
			err error,
		) {
			// defer func() {
			// 	fmt.Println(
			// 		"marshal result: in:", v,
			// 		"data:", string(data),
			// 		"type:", payloadType,
			// 		"err:", err,
			// 	)
			// }()
			x, ok := v.([]byte)
			if !ok {
				return nil, 0, errors.New("codec: invalid marshal input")
			}
			return x, websocket.BinaryFrame, nil
		},
		Unmarshal: func(
			data []byte,
			payloadType byte,
			v interface{},
		) (err error) {
			// fmt.Println(
			// 	"v type1:", reflect.ValueOf(v).Kind(),
			// )
			// defer func() {
			// 	fmt.Println(
			// 		"v type2:", reflect.ValueOf(v).Kind(),
			// 	)
			// 	fmt.Println(
			// 		"unmarshal result: in:", string(data),
			// 		"v:", v,
			// 		"type:", payloadType,
			// 		"err:", err,
			// 	)
			// }()
			if payloadType != websocket.BinaryFrame {
				return errors.New("codec: invalid unmarshal input")
			}
			var redirect *[]byte
			redirect, ok := v.(*[]byte)

			if !ok {
				return errors.New("codec: invalid unmarshal input: can't redirect v as *[]byte")
			}

			*redirect = data

			return nil
		},
	}
	return codec
}

func (self *Controller) HandleARPCCall(
	call *goarpcsolution.ARPCCall,
) (error, error) {
	return nil, nil
}

func main() {

	ctl := &Controller{}

	router := httprouter.New()

	router.GET("/hello", ctl.HelloPage)
	router.POST("/hello", ctl.HelloPage)

	router.Handler(
		"GET",
		"/wasm_example",
		http.RedirectHandler("/wasm_example/", 301),
	)

	router.HandlerFunc(
		"GET",
		"/ws_jrpc2",
		func(
			rw http.ResponseWriter,
			rq *http.Request,
		) {
			s := &websocket.Server{}
			s.Handler = ctl.WSHandlerJRPC2
			s.ServeHTTP(rw, rq)
		},
	)

	router.HandlerFunc(
		"GET",
		"/ws_arpc",
		func(
			rw http.ResponseWriter,
			rq *http.Request,
		) {
			s := &websocket.Server{}
			s.Handler = ctl.WSHandlerARPC
			s.ServeHTTP(rw, rq)
		},
	)

	router.Handle(
		"GET",
		"/wasm_example/*path",
		func(
			rw http.ResponseWriter,
			rq *http.Request,
			params httprouter.Params,
		) {
			// params := httprouter.ParamsFromContext(r.Context())
			p := params.ByName("path")
			p = strings.ReplaceAll(p, "/", "")
			p = strings.Trim(p, ".")
			http.ServeFile(rw, rq, "./static/"+p)
		},
	)

	certificate, err := tls.LoadX509KeyPair("./tls/cert.pem", "./tls/key.pem")
	if err != nil {
		log.Fatal("error loading keyfile or certificate: ", err)
	}

	tls_cfg := &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: true,
	}

	addr := ":8080"

	log.Println("address:", addr)

	s := &http.Server{
		Addr:      addr,
		Handler:   router,
		TLSConfig: tls_cfg,
	}

	log.Fatal(s.ListenAndServeTLS("", ""))

}
