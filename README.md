# Pebble client

Go client library for interacting with Pebble-style apps.

# Usage

## Creating client

```go
import "github.com/t11e/go-pebbleclient"

func main() {
  client, err := New(ClientOptions{
    AppName: "central",
    ApiVersion: 1,
    Session: "uio3uio43uio4oui432",
    Host: "localhost",
  })
}
```
