package main

import (
	"fmt"
	"log"
  "bytes"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"

  "encoding/json"

  "context"
  "./whoislocal"
  "./dbsetup"
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
    url := "https://api.ssllabs.com/api/v3/analyze?host=" + string(name)
    fmt.Fprintf(ctx, "Url = %s\n", string(url))
  } else {
    fmt.Fprintf(ctx, "Please specify a domain!\n")
  }
}

func main() {
  router := fasthttprouter.New()
  router.GET("/", Index)
  router.GET("/domains", QueryArgs)

  fmt.Println("server starting on localhost:3000")

  log.Fatal(fasthttp.ListenAndServe(":3000", router.Handler))
}