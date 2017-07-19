package main

import (
        "fmt"
	"time"
)

func main() {
        i := 0
	k := true
        for {
                i++
                if k == false{
                        break
                }
                fmt.Printf("%v ", i)
		time.Sleep(time.Millisecond * 10)
        }
        fmt.Println()
}
