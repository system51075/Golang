package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"golang.org/x/net/proxy"
	"flag"
)
var P_proxy = flag.String("proxy", "socks5://10.30.22.156:8082", "a string")
var P_site = flag.String("site", "http://showip.net", "a string")
func ProxyAwareHttpClient() *http.Client {
	// sane default
	var dialer proxy.Dialer
	dialer = proxy.Direct
	proxyServer := *P_proxy
	isSet := true
//	fmt.Println(isSet)
//	fmt.Println(proxyServer)
	if isSet {
		proxyUrl, err := url.Parse(proxyServer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid proxy url %q\n", proxyUrl)
		}
		dialer, err = proxy.FromURL(proxyUrl, proxy.Direct)
	}

	// setup a http client
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	return httpClient
}

func main() {
	flag.Parse()
	fmt.Print("Use: ./socks5 --proxy=socks5://10.30.22.156:8082 --site=http://showip.net", "\n")
	fmt.Print("Auth: toanpt3@vng\n")
	fmt.Print("---------------------------------------------\n")
	fmt.Print("Socks: ",*P_proxy, "  ---> ", *P_site, "\n")
	fmt.Print("\n --------------- REsult--------------------\n")
	req, err := http.NewRequest("GET", *P_site, nil)
	if err != nil {
		panic(err)
	}

	client := ProxyAwareHttpClient()
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(contents))
}
