# Pebble client

Go client library for interacting with Pebble-style apps.

# Usage

## Creating client

```go
import (
  pc "github.com/t11e/go-pebbleclient"
)

func main() {
  builder, err := pc.NewHTTPClientBuilder(pc.Options{
    ServiceName: "central",
    ApiVersion: 1,
    Session: "uio3uio43uio4oui432",
    Host: "localhost",
  })
  if err != nil {
    log.Fatal(err)
  }
  client, err := builder.NewClient(pc.ClientOptions{})
  if err != nil {
    log.Fatal(err)
  }
  // ...
}
```

Or from request:

```go
func myHandler(w http.ResponseWriter, req *http.Request) {
  client, err := builder.NewClient(pc.ClientOptions{}).FromHTTPRequest(req)
  // ...
}
```

## `GET` requests

```go
var result *Organization
err := client.Get("/posts/post.listing:endeavor.enclosure.missoulian$1178983", &pc.RequestOptions{
  Params: pc.Params{"raw": true},
}, &result)
```

## `HEAD` requests

```go
err := client.Head("/posts/post.listing:endeavor.enclosure.missoulian$1178983", nil)
```

## `POST` requests

```go
b, err := json.Marshal(organization)
var result *Organization
err := client.Post("/organizations", nil, bytes.NewReader(b)}, &result)
```

## `PUT` requests

```go
b, err := json.Marshal(organization)
var result *Organization
err := client.Put("/organizations/1", nil, bytes.NewReader(b)}, &result)
```

## `DELETE` requests

```go
err := client.Delete("/organizations/1", nil, nil, nil)
```
