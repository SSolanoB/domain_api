package main

import (
	"fmt"
	"log"
  "strings"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"

  //"encoding/json"

  "./dbsetup"
  "./dbsetup/ssllabsapi"
)

func Index(ctx *fasthttp.RequestCtx) {
  fmt.Fprintf(ctx, "Welcome!\n")
}

func Hello(ctx *fasthttp.RequestCtx) {
  fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}

func QueryArgs(ctx *fasthttp.RequestCtx) {
  name := ctx.QueryArgs().Peek("name")
  fmt.Fprintf(ctx, "Pong! %s\n", string(name))
  if name != nil {
    str_name := string(name)
    if strings.HasPrefix(str_name, "http") != true {
      str_name = "http://" + str_name
    }
    url := "https://api.ssllabs.com/api/v3/analyze?host=" + str_name
    fmt.Fprintf(ctx, "Url = %s\n", string(url))
    respon := ssllabsapi.RequestApi(string(url))
    // Should wait until api respond with info of servers.
    dbsetup.ExecuteTransaction(respon)
  } else {
    fmt.Fprintf(ctx, "Please specify a domain!\n")
  }
}

func main() {
  router := fasthttprouter.New()
  router.GET("/", Index)
  router.GET("/domains", QueryArgs)


  fmt.Println("server starting on localhost:3000")

  dbsetup.SetupDb()

  log.Fatal(fasthttp.ListenAndServe(":3000", router.Handler))
}