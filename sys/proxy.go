package main
 
import (
   "fmt"
   "net/http"
   "net/url"
   "io/ioutil"
)
 
func main() {
 
   proxyString := "10.30.58.36:81"
   proxyURL, err := url.Parse(proxyString)
   checkError(err)
 
   rawURL := "http://showip.net"
   url, err := url.Parse(rawURL)
   checkError(err)
 
   transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
   client := &http.Client{Transport: transport}
 
   request, err := http.NewRequest("GET", url.String(), nil)
   checkError(err)
 
   response, err := client.Do(request)
   checkError(err)
 
   htmlData, err := ioutil.ReadAll(response.Body)
   checkError(err)
   fmt.Println(string(htmlData))
 
}
 
func checkError(err error) {
   if err != nil {
      fmt.Println(err)
   }
}
 
