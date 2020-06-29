package ssllabsapi

import (
  "fmt"
  "bytes"

  "github.com/valyala/fasthttp"

  "encoding/json"
)


type Response struct {
  Host string
  Port int
  Protocol string
  IsPublic bool
  Status string
  StartTime int
  TestTime int
  EngineVersion string
  CriteriaVersion string
  Endpoints []Response2
}
//change default. To nil! 
// validations! 
type Response2 struct {
  IpAddress string
  ServerName string
  StatusMessage string
  Grade string
  GradeTrustIgnored string
  HasWarnings bool
  IsExceptional bool
  Progress int
  Duration int
  Delegation int
}

func RequestApi(url string) Response {
	req := fasthttp.AcquireRequest()
  defer fasthttp.ReleaseRequest(req)
  req.SetRequestURI(url)

  resp := fasthttp.AcquireResponse()
  defer fasthttp.ReleaseResponse(resp)

	fmt.Printf("URL is: %s\n", url)  
  
  err := fasthttp.Do(req, resp)
  if err != nil {
    fmt.Printf("Request failed: %s\n", err)
    var r Response
    return r
  }
  if resp.StatusCode() != fasthttp.StatusOK {
    fmt.Printf("Bad satus code %d\n", resp.StatusCode())
    var r Response
    return r
  }

  contentType := resp.Header.Peek("Content-Type")
  fmt.Printf("Content type is: %s\n", contentType)

  body := resp.Body()

  dec := json.NewDecoder(bytes.NewReader(body))

  var r Response

  rerr := dec.Decode(&r)

  if rerr != nil {
    fmt.Printf("Request failed inside: %s\n", rerr)
    var r Response
    return r
  }

  fmt.Printf("Body is: %T\n", body)
  return r
}