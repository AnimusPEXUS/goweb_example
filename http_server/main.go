package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/AnimusPEXUS/gojsonrpc2"
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

	codec := &websocket.Codec{}

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

	wg.Add(1)

	go func() {
		defer wg.Done()
		var data []byte
		for true {
			if stop_flag {
				return
			}
			data = make([]byte, 0)
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

	wg.Wait()

	fmt.Println("rec_err:", rec_err)
	fmt.Println("send_err:", send_err)
	fmt.Println("rec_proto_err:", rec_proto_err)

}

// func (self *Controller) WasmExamplePage(
// 	rw http.ResponseWriter,
// 	rq *http.Request,
// 	// args map[string]string,
// ) {
// 	http.ServeFile(rw, rq, "./static/wasm.html")
// 	return
// }

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

	s := &http.Server{
		Addr:      ":8080",
		Handler:   router,
		TLSConfig: tls_cfg,
	}

	s.ListenAndServe()

}
