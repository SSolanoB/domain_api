package ssllabsapi

import (
  "fmt"

  "github.com/buaazp/fasthttprouter"
  "github.com/valyala/fasthttp"

  "encoding/json"
)


func RequestApi(url string) response {
	req := fasthttp.AcquireRequest()
  defer fasthttp.ReleaseRequest(req)
  req.SetRequestURI(url)

  resp := fasthttp.AcquireResponse()
  defer fasthttp.ReleaseResponse(resp)

	fmt.Printf("URL is: %s\n", url)  
  
  err := fasthttp.Do(req, resp)
  if err != nil {
    fmt.Printf("Request failed: %s\n", err)
    var r response
    return r
  }
  if resp.StatusCode() != fasthttp.StatusOK {
    fmt.Printf("Bad satus code %d\n", resp.StatusCode())
    var r response
    return r
  }

  contentType := resp.Header.Peek("Content-Type")
  fmt.Printf("Content type is: %s\n", contentType)

  body := resp.Body()

  dec := json.NewDecoder(bytes.NewReader(body))

  var r response

  rerr := dec.Decode(&r)

  if rerr != nil {
    fmt.Printf("Request failed inside: %s\n", rerr)
    var r response
    return r
  }

  fmt.Printf("Body is: %T\n", body)
  return r
}