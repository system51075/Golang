package main

  import (
    "fmt"
    "github.com/akamai-open/AkamaiOPEN-edgegrid-golang"
    "io/ioutil"
	"bytes"
    "net/http"
  )

  func main() {
    client := http.Client{}
    var jsonStr = []byte(`{"objects": ["http://znews-photo-akamai.zadn.vn/w660/Uploaded/kpcwvovd/2017_07_02/1_181960.jpg"]}`)
//    config, _ := edgegrid.Init("./data", "default")
  config := edgegrid.Config{
      Host : "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX.purge.akamaiapis.net",
      ClientToken:  "XXXXXXXXXXXXXXXXXXXXXXXXXXX",
      ClientSecret: "XXXXXXXXXXXXXXXXXXXXXXXXX=",
      AccessToken:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",      
      MaxBody:      1024,
      //HeaderToSign: []string{
      //  "X-Test1",
      //  "X-Test2",
      //  "X-Test3",
      //},
      Debug:        false,
    }  
    // Retrieve all locations for diagnostic tools
//    req, _ := http.NewRequest("POST", fmt.Sprintf("https://%s/ccu/v3/delete/url/staging", config.Host), bytes.NewBuffer(jsonStr))
    req, _ := http.NewRequest("POST", fmt.Sprintf("https://%s/ccu/v3/delete/url/production", config.Host), bytes.NewBuffer(jsonStr))
    req = edgegrid.AddRequestHeader(config, req)
    resp, _ := client.Do(req)

    defer resp.Body.Close()
    byt, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(byt))
  }
