package main

import (
	"fmt"
	"html"
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

type Controller struct {
	Name string
}

func (self *Controller) HelloPage(
	rw http.ResponseWriter,
	rq *http.Request,
	params map[string]string,
) {
	var name_part string

	name_part = "<noname>"
	if self.Name != "" {
		name_part = self.Name
	}

	additional_message := ""

	switch rq.Method {
	default:
		additional_message = fmt.Sprintf("unsupported method: %s", rq.Method)
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
		additional_message = fmt.Sprintf("(Used method is %s)", rq.Method)
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

func main() {

	ctl := &Controller{}

	router := httptreemux.New()
	router.GET("/hello", ctl.HelloPage)
	router.POST("/hello", ctl.HelloPage)

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	s.ListenAndServe()

}
