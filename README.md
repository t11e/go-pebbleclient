# Pebble client

Go client library for interacting with Pebble-style apps.

# Usage

## Creating client

```go
import "github.com/t11e/go-pebbleclient"

func main() {
  client, err := pebbleclient.New(pebbleclient.ClientOptions{
    AppName: "central",
    ApiVersion: 1,
    Session: "uio3uio43uio4oui432",
    Host: "localhost",
  })
}
```

Or from request:

```go
func myHandler(w http.ResponseWriter, req *http.Request) {
  client, err := pebbleclient.NewFromRequest(pebbleclient.ClientOptions{
    AppName: "central",
  }, req)
  // ...
}
```

## `GET` requests

```go
var result *Organization
client.Get("/organization/1", pebbleclient.Params{
  "session": "smurf42",
}, &result)
```

## `POST` requests

```go
b, err := json.Marshal(organization)
var result *Organization
client.Post("/organizations", pebbleclient.Body{Data: bytes.NewReader(b)}, &result)
```
