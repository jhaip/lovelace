package main

import (
  "time"
  "fmt"
)

var N int = 100
var i int = N

func ping() {
  pong()
}

func pong() {
  i -= 1
  if i > 0 {
    ping()
  }
}

func main() {
  start := time.Now()
  ping()
  end := time.Since(start)
  fmt.Printf("total     : %s \n", end)
}
