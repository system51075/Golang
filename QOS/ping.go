package main

import (
    "fmt"
    "regexp"
    "strings"
    "strconv"
)

func main() {
    strOutput := `1->5 packets transmitted, 5 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 0.067/0.078/0.087/0.007 ms 
        2->5 packets transmitted, 5 received, 0% packet loss, time 801ms
rtt min/avg/max/stddev = 0.019/0.034/0.044/0.010 ms, ipg/ewma 200.318/0.038 ms`
    latencyPattern := regexp.MustCompile(`(round-trip|rtt)\s+\S+\s*=\s*([0-9.]+)/([0-9.]+)/([0-9.]+)/([0-9.]+)\s*ms`)
    matches := latencyPattern.FindAllStringSubmatch(strOutput, -1)
    for _, item := range matches {
        latency, _ := strconv.ParseFloat(strings.TrimSpace(item[3]), 64)
            jitter, _ := strconv.ParseFloat(strings.TrimSpace(item[5]), 64)
            fmt.Printf("AVG = %.3f, STDDEV = %.3f\n", latency, jitter)

        }
}

