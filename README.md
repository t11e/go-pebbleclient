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

# Contributions

Clone this repository into your GOPATH (`$GOPATH/src/github.com/t11e/`)
and use [Glide](https://github.com/Masterminds/glide) to install its dependencies.

```sh
brew install glide
go get github.com/t11e/go-pebbleclient
cd "$GOPATH"/src/github.com/t11e/go-pebbleclient
glide install --strip-vendor
```

You can then run the tests:

```sh
go test $(go list ./... | grep -v /vendor/)
```

There is no need to use `go install` as any project that requires this library
can include it as a dependency like so:

```sh
cd my_other_project
glide get --strip-vendor github.com/t11e/go-pebbleclient
```

If you change any of the interfaces that have a mock in `mocks/` directory be sure to execute
`go generate` and check in the updated mock files.
