# fnos-tunnel (Go)

fnOS Cloudflare Tunnel SDK for Go.

## 安装

```bash
go get github.com/dustink66/fnos-tunnel-sdk-go
```

## 快速开始

```go
package main

import (
    "fmt"
    "time"
    tunnel "github.com/dustink66/fnos-tunnel-sdk-go"
)

func main() {
    client := tunnel.NewAPIClient(
        "http://<your-fnos-ip>:19092",
        "<your_app_id>",
        "<your_app_key>",
        "com.example.myapp",
        10*time.Second,
    )

    fmt.Println(client.Health())
    status, _ := client.Status()
    fmt.Printf("Running: %v\n", status.Running)
    ds, _ := client.DomainStatus("")
    fmt.Printf("Registered: %v\n", ds.Registered)
    result, _ := client.Register("myapp.example.com", "http://localhost:8080")
    fmt.Printf("Success: %v\n", result.Success)
}
```