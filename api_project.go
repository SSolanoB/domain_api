package main

import (
	"fmt"
	"log"
  "strings"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"

  "encoding/json"

  "./dbsetup"
  "./dbsetup/ssllabsapi"
)

func Index(ctx *fasthttp.RequestCtx) {
  fmt.Fprintf(ctx, "Welcome!\n")
}

func Hello(ctx *fasthttp.RequestCtx) {
  fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}

func DomainIndex(ctx *fasthttp.RequestCtx) {
  //fmt.Fprintf(ctx, "I will give you the domains!\n")
  response, err := dbsetup.ReturnDomains()
  fmt.Println(response)
  if err != nil {
    ctx.SetStatusCode(fasthttp.StatusBadRequest)
    ctx.SetBodyString("There was an error, try it again later!")
  } else {
    enc := json.NewEncoder(ctx)
    err = enc.Encode(&response)
    ctx.SetStatusCode(fasthttp.StatusOK)
  }
  
  ctx.SetContentType("application/json")
  ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
  ctx.Response.Header.Set("Access-Control-Allow-Headers", "authorization")
  ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET")
  ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
}

func QueryArgs(ctx *fasthttp.RequestCtx) {
  name := ctx.QueryArgs().Peek("name")
  //fmt.Fprintf(ctx, "Pong! %s\n", string(name))
  if name != nil {
    str_name := string(name)
    if strings.HasPrefix(str_name, "http") != true {
      str_name = "http://" + str_name
    }
    url := "https://api.ssllabs.com/api/v3/analyze?host=" + str_name
    //fmt.Fprintf(ctx, "Url = %s\n", string(url))
    respon := ssllabsapi.RequestApi(string(url))
    // Should wait until api respond with info of servers.
    r, err := dbsetup.ExecuteTransaction(respon)

    fmt.Println(r.Title)

    if err != nil {
      fmt.Fprintf(ctx, "There was an error!")
      ctx.SetStatusCode(fasthttp.StatusBadRequest)
    } else {
      enc := json.NewEncoder(ctx)
      enc.Encode(&r)
      ctx.SetStatusCode(fasthttp.StatusOK)
    }
    ctx.SetContentType("application/json")
    ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
    ctx.Response.Header.Set("Access-Control-Allow-Headers", "authorization")
    ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET")
    ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
  } else {
    fmt.Fprintf(ctx, "Please specify a domain!\n")
    ctx.SetStatusCode(fasthttp.StatusBadRequest)
  }
}

func main() {
  router := fasthttprouter.New()
  router.GET("/", Index)
  router.GET("/domain", QueryArgs)
  router.GET("/domains", DomainIndex)


  fmt.Println("server starting on localhost:3000")

  dbsetup.SetupDb()

  log.Fatal(fasthttp.ListenAndServe(":3000", router.Handler))
}