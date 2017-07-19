package main

import (
	"flag"
	"fmt"
	influxdbclient  "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/client/v2"

	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	 _ "github.com/influxdata/influxdb/client/v2"
	"math/rand"
)
const (
    MyDB = "qos"
    username = ""
    password = ""
)

var (
	receiverHostFlag     string // receiverHost: Host running statistics daemon process
	receiverPortFlag     int    // receiverPort: Port running statistics daemon process
	receiverNoopFlag     bool   // If true, do not actually send the metrics to the receiver.
	receiverTypeFlag     string // receiverType: Set type of receiver for statistics
	receiverUsernameFlag string // receiverUsername: Optional username string for InfluxDB receiver.
	receiverPasswordFlag string // receiverPassword: Optional password string for InfluxDB receiver.
	receiverDatabaseFlag string // receiverDatabase: Database database string for InfluxDB receiver.

	hostsFlag          string // Flag hosts: Comma-seperated list of IP/hostnames to ping
	pingCountFlag      uint64 // Flag count: Uint8 Interger number of pings to send per cycle.
	oneshotFlag        bool
	originFlag         string // String denoting specific origin hostname used in metric submission.
	intervalFlag       time.Duration
	verboseFlag        bool
	re_ping_packetloss *regexp.Regexp
	re_ping_rtt        *regexp.Regexp
	re_ping_hostname   *regexp.Regexp
	quietFlag          bool

	pingBinary string // Path to ping binary based upon operating system)

)

type PingStats struct {
	loss float64
	min  float64
	avg  float64
	max  float64
	mdev float64
}

type Ping struct {
	origin      string
	destination string
	time        int64
	stats       PingStats
}

func isReceiverFullyDefined() bool {
	if len(receiverTypeFlag) == 0 {
		// No receiver is defined, so don't try to judge its validity.
		return true
	}
	if len(receiverHostFlag) > 0 && (receiverPortFlag > 0) {
		return true
	}
	return false
}

func checkValidReceiverType(rType string, validTypes []string) bool {
	if rType == "" {
		return true
	}
	for _, v := range validTypes {
		if rType == v {
			return true
		}
	}
	return false
}


func setOsParams(os string) {
	re_ping_hostname = regexp.MustCompile(`--- (?P<hostname>\S+) ping statistics ---`)

	switch runtime.GOOS {
	case "openbsd":
		pingBinary = "/sbin/ping"
		re_ping_packetloss = regexp.MustCompile(`(?P<loss>\d+.\d+)\% packet loss`)
		re_ping_rtt = regexp.MustCompile(`round-trip min/avg/max/std-dev = (?P<min>\d+.\d+)/(?P<avg>\d+.\d+)/(?P<max>\d+.\d+)/(?P<mdev>\d+.\d+) ms`)
	case "linux":
		pingBinary = "/bin/ping"
		re_ping_packetloss = regexp.MustCompile(`(?P<loss>\d+)\% packet loss`)
		distro := os
		switch distro {
		case "ubuntu", "debian":
			re_ping_rtt = regexp.MustCompile(`(rtt|round-trip) min/avg/max/(mdev|stddev) = (?P<min>\d+.\d+)/(?P<avg>\d+.\d+)/(?P<max>\d+.\d+)/(?P<mdev>\d+.\d+) ms`)
		case "alpine":
			re_ping_rtt = regexp.MustCompile(`round-trip min/avg/max = (?P<min>\d+.\d+)/(?P<avg>\d+.\d+)/(?P<max>\d+.\d+) ms`)
		}
	default:
		log.Fatalf("Unsupported operating system. runtime.GOOS: %v.\n", runtime.GOOS)
	}
}

func init() {
	flag.StringVar(&hostsFlag, "hosts", "", "Comma-seperated list of hosts to ping.")
	flag.Uint64Var(&pingCountFlag, "pingcount", 5, "Number of pings per cycle.")
	flag.BoolVar(&oneshotFlag, "oneshot", false, "Execute just one ping round per host. Do not loop.")
	flag.StringVar(&originFlag, "origin", "", "Override hostname as origin with this value.")
	flag.DurationVar(&intervalFlag, "interval", 60*time.Second, "Seconds of wait in between each round of pings.")
	flag.StringVar(&receiverHostFlag, "receiverhost", "", "Hostname of metrics receiver. Optional")
	flag.IntVar(&receiverPortFlag, "receiverport", 0, "Port of receiver.")
	flag.BoolVar(&receiverNoopFlag, "receivernoop", false, "If set, do not send Metrics to receiver.")
	flag.BoolVar(&verboseFlag, "v", false, "If set, print out metrics as they are processed.")
	flag.StringVar(&receiverUsernameFlag, "receiverusername", "", "Username for InfluxDB database. Optional.")
	flag.StringVar(&receiverPasswordFlag, "receiverpassword", "", "Password for InfluxDB database. Optional.")
	flag.StringVar(&receiverDatabaseFlag, "receiverdatabase", "", "Database for InfluxDB.")
	flag.BoolVar(&quietFlag, "q", false, "If set, only log in case of errors.")
}

// Return true if host resolves, false if not.
func doesHostExist(host string) bool {
	addresses, _ := net.LookupHost(host)
	if len(addresses) > 0 {
		return true
	}
	return false
}

func getValidHosts(hosts []string) []string {
	var trimmedHosts []string
	for _, currentHost := range hosts {
		if doesHostExist(currentHost) {
			trimmedHosts = append(trimmedHosts, currentHost)
		}
	}
	return trimmedHosts
}

func processPingOutput(pingOutput string, pingErr bool) Ping {
	var ping Ping
	var stats PingStats
	now := time.Now()
	ping.time = now.Unix()
	if len(originFlag) == 0 {
		origin, _ := os.Hostname()
		ping.origin = origin
	} else {
		ping.origin = originFlag
	}

	re_ping_hostname_matches := re_ping_hostname.FindAllStringSubmatch(pingOutput, -1)[0]
	ping.destination = re_ping_hostname_matches[1]

	re_packetloss_matches := re_ping_packetloss.FindAllStringSubmatch(pingOutput, -1)[0]

	stats.loss, _ = strconv.ParseFloat(re_packetloss_matches[1], 64)

	if pingErr == true {
		stats.min, stats.avg, stats.max, stats.mdev = 0, 0, 0, 0
	} else {
		re_rtt_matches := re_ping_rtt.FindAllStringSubmatch(pingOutput, -1)[0]
		rtt_map := make(map[string]string)
		for i, name := range re_ping_rtt.SubexpNames() {
			if i != 0 {
				rtt_map[name] = re_rtt_matches[i]
			}
		}
		stats.min, _ = strconv.ParseFloat(rtt_map["min"], 64)
		stats.avg, _ = strconv.ParseFloat(rtt_map["avg"], 64)
		stats.max, _ = strconv.ParseFloat(rtt_map["max"], 64)
		stats.mdev, _ = strconv.ParseFloat(rtt_map["mdev"], 64)
	}
	ping.stats = stats
	return ping
}

func executePing(host string, numPings uint64) (string, bool) {
	pingError := false
	countFlag := fmt.Sprintf("-c%v", numPings)
	out, err := exec.Command(pingBinary, countFlag, host).Output()
	if err != nil {
		log.Printf("Error with host %s, error: %s, output: %s.\n", host, err, out)
		pingError = true
	}
	s_out := string(out[:])
	if verboseFlag {
		log.Printf("Raw Ping Output: %v.\n", s_out)
	}
	return s_out, pingError
}

func spawnPingLoop(c chan<- Ping,
	host string,
	numPings uint64,
	sleepTime time.Duration,
	oneshot bool) {
	for {
		raw_output, err := executePing(host, numPings)
		pingResult := processPingOutput(raw_output, err)
		c <- pingResult
		time.Sleep(sleepTime)

		if oneshot == true {
			break
		}
	}
}


func createInfluxDBMetrics(ping Ping) (influxdbclient.BatchPoints, error) {
	var err error
	bp, err := influxdbclient.NewBatchPoints(influxdbclient.BatchPointsConfig{
		Database:  receiverDatabaseFlag,
		Precision: "s",
	})
	if err != nil {
		return nil, err
	}

	tags := map[string]string{
		"origin":      ping.origin,
		"destination": ping.destination,
	}
	fields := map[string]interface{}{
		"loss": ping.stats.loss,
		"min":  ping.stats.min,
		"avg":  ping.stats.avg,
		"max":  ping.stats.max,
		"mdev": ping.stats.mdev,
	}
	//pt, err := influxdbclient.NewPoint("ping", tags, fields, time.Now())
	pt, err := influxdbclient.NewPoint("ping", tags, fields, time.Now())
	//fmt.Println(pt)
	if err != nil {
		return nil, err
	}
	
	bp.AddPoint(pt)
	return bp, nil
}

func processPing(c <-chan Ping) error {
	var err error

	//var ic influxdbclient.Client

	if err != nil {
		log.Printf("Error in creating connection to receiver: %v.\n", err)
	}
	for {
		pingResult := <-c
		if !isReceiverFullyDefined() {
			// A receiver is not fully defined.
			log.Printf("You receiver is not fully defined. Host: %v, Port: %v.\n", receiverHostFlag, receiverPortFlag)
			continue
		}
		//fmt.Println(pingResult)
			_, err := createInfluxDBMetrics(pingResult)
		//fmt.Println(a)
			if err != nil {
				log.Fatalln(err)
			}


	}
	return nil
}

func main() {
	x, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: "http://127.0.0.1:8086",
        Username: username,
        Password: password,
    })
	distro := "ubuntu"
	flag.Parse()
	setOsParams(distro)
	rand.Seed(42)

    bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
        Database: MyDB,
        Precision: "us",
    })
    if err != nil {
        log.Fatalln("Error: ", err)
    }
	hasValidReceiver := checkValidReceiverType(receiverTypeFlag, []string{"carbon", "influxdb"})
	if !hasValidReceiver {
		log.Fatalf("You specified an unsupported receiver type %v.\n", receiverTypeFlag)
	}
	hosts := strings.Split(hostsFlag, ",")
	validHosts := getValidHosts(hosts)
	
	var c chan Ping = make(chan Ping)
	for _, currentHost := range validHosts {
		go spawnPingLoop(c, currentHost, pingCountFlag, intervalFlag, oneshotFlag)
	}
		for {

		pingResult := <-c
		if !isReceiverFullyDefined() {
			// A receiver is not fully defined.
			log.Printf("You receiver is not fully defined. Host: %v, Port: %v.\n", receiverHostFlag, receiverPortFlag)
			continue
		}

         tags := map[string]string{
		"origin":      pingResult.origin,
		"destination": pingResult.destination,
	}
	fields := map[string]interface{}{
		"loss": pingResult.stats.loss,
		"min":  pingResult.stats.min,
		"avg":  pingResult.stats.avg,
		"max":  pingResult.stats.max,
		"mdev": pingResult.stats.mdev,
	}
        pt, err := client.NewPoint("ping", tags, fields, time.Now())
        if err != nil {
            log.Fatalln("Error: ", err)
        }
        bp.AddPoint(pt)
        time.Sleep(time.Millisecond * 10)
    x.Write(bp)
		


	}
}
