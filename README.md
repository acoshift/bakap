# bakap

Backup files to google cloud storage

## Usage

```go
package main

import (
  "time"

  "github.com/acoshift/bakap"
)

func main() {
  bakap.Start(bakap.Config{
    Interval: time.Hour * 24,
    Files: []bakap.File{
      // upload from "appendonly.aof" to "[Time] redis-6379.aof"
      bakap.File{Src: "/var/lib/redis/6379/appendonly.aof", Dest: "redis-6379.aof"},
    },
    Bucket:  "BUCKET-NAME",
    Account: "xxx@project-id.iam.gserviceaccount.com",
    PrivateKey: []byte(`PRIVATE KEY for xxx@project-id.iam.gserviceaccount.com`),
  })
}
```
