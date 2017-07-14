package main

import (
"fmt"
"log"
"time"
"math/rand"

"github.com/influxdata/influxdb/client/v2"
)

const (
    MyDB = "qos"
    username = ""
    password = ""
)

func main(){
    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: "http://127.0.0.1:8086",
        Username: username,
        Password: password,
    })

    if err != nil {
        log.Fatalln("Error: ", err)
    }

    writePoints(c, MyDB)
}

func writePoints(c client.Client, MyDB string) {
    sampleSize := 1000
    rand.Seed(42)

    bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
        Database: MyDB,
        Precision: "us",
    })

    for i := 0; i < sampleSize; i++ {
        regions := []string{"us-west1", "us-west2", "us-west3", "us-east1"}
        tags := map[string]string{
            "cpu": "cpu-total",
            "host": fmt.Sprintf("host%d", rand.Intn(1000)),
            "region": regions[rand.Intn(len(regions))],
        }

        idle := rand.Float64() * 100.0
        fields := map[string]interface{}{
            "idle": idle,
            "busy": 100.0 - idle,
        }

        pt, err := client.NewPoint("cpu_usage", tags, fields, time.Now())
        fmt.Println(pt)
        if err != nil {
            log.Fatalln("Error: ", err)
        }

        bp.AddPoint(pt)
    }

    err := c.Write(bp)
    if err != nil {
        log.Fatal(err)
    }
}