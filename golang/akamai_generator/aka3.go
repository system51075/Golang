package main
// Akamai generator 
// Use: ./aka3 --verbose --end_time=$TIMESTAMP --key=$KEY --algo=md5 --url=$URL --token_name=authen
// Example: ./aka3 --verbose --end_time=1500024025 --key=abc123 --algo=md5 --url=/7074edb5d0f039ae60e1/4285108012442572695 --token_name=authen
//FUll url : http://SITE/7074edb5d0f039ae60e1/4285108012442572695?authen=exp=1500024025~hmac=71128bd5230de8c4b24f540560f2635e 
import (

	"bytes"

	"crypto/hmac"

	"crypto/md5"

	"crypto/sha1"

	"crypto/sha256"

	"encoding/hex"

	"flag"

	"fmt"

	"io"

	"log"

	"net/http"

	"net/url"

	"os"

	"regexp"

	"strconv"

	"strings"

	"time"

)


func checkError(err error, msg string) {

	if err != nil {

		//log.Fatal(err)

		if msg != "" {

			panic(msg)

		} else {

			panic(err)

		}

	}

}


type TokenParams struct {

	token_type         string

	token_name         string

	ip                 string

	start_time         string

	end_time           string

	window_seconds     string

	url                string

	acl                string

	key                string

	payload            string

	algo               string

	salt               string

	session_id         string

	field_delimiter    string

	acl_delimiter      string

	escape_early       bool

	escape_early_upper bool

	verbose            bool

}


func escapeEarly(token_params *TokenParams, value string) string {

	if token_params.escape_early || token_params.escape_early_upper {

		hex_regex := regexp.MustCompile("%..")

		if token_params.escape_early_upper {

			// Replace all instances of %xx with %XX

			return string(hex_regex.ReplaceAllFunc(

				[]byte(url.QueryEscape(value)),

				func(s []byte) []byte { return []byte(strings.ToUpper(string(s))) }))

		} else if token_params.escape_early {

			// Replace all instances of %XX with %xx

			return string(hex_regex.ReplaceAllFunc(

				[]byte(url.QueryEscape(value)),

				func(s []byte) []byte { return []byte(strings.ToLower(string(s))) }))

		}

	}

	return value

}


func GenerateToken(token_params *TokenParams, w io.Writer) string {

	if token_params.token_name == "" {

		token_params.token_name = "hdntl"

	}

	if token_params.algo == "" {

		token_params.algo = "sha256"

	}

	if token_params.field_delimiter == "" {

		token_params.field_delimiter = "~"

	}

	if token_params.acl_delimiter == "" {

		token_params.acl_delimiter = "!"

	}

	var numeric_start_time int64

	if strings.ToLower(token_params.start_time) == "now" {

		numeric_start_time = time.Now().Unix()

	} else if token_params.start_time != "" {

		start_time, err := strconv.ParseInt(token_params.start_time, 10, 64)

		checkError(err, "start_time must be numeric or now.")

		numeric_start_time = start_time

	}

	var numeric_end_time int64

	if token_params.end_time != "" {

		end_time, err := strconv.ParseInt(token_params.end_time, 10, 64)

		checkError(err, "end_time must be numeric.")

		numeric_end_time = end_time

	}

	var numeric_window_seconds int64

	if token_params.window_seconds != "" {

		window_seconds, err := strconv.ParseInt(token_params.window_seconds, 10, 64)

		checkError(err, "window_seconds must be numeric.")

		numeric_window_seconds = window_seconds

	}

	if numeric_end_time <= 0 {

		if numeric_window_seconds > 0 {

			if token_params.start_time == "" {

				// If we have a duration window without a start time,

				// calculate the end time starting from the current time.

				numeric_end_time = time.Now().Unix() + numeric_window_seconds

			} else {

				numeric_end_time = numeric_start_time + numeric_window_seconds

			}

		} else {

			panic("You must provide an expiration time or a duration window.")

		}

	}

	if numeric_end_time < numeric_start_time {

		fmt.Fprintln(w, "WARNING:Token will have already expired.")

	}

	if token_params.key == "" {

		panic("You must provide a secret in order to generate a new token.")

	}

	if token_params.acl == "" && token_params.url == "" {

		panic("You must provide an ACL OR a URL.")

	}

	if token_params.acl != "" && token_params.url != "" {

		panic("You must provide an ACL OR a URL, not both.")

	}

	if token_params.verbose {

		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "Akamai Token Generator v2.0.7")

		fmt.Fprintln(w, "token_type         :", token_params.token_type)

		fmt.Fprintln(w, "token_name         :", token_params.token_name)

		fmt.Fprintln(w, "algorithm          :", token_params.algo)

		fmt.Fprintln(w, "start_time         :", numeric_start_time)

		fmt.Fprintln(w, "end_time           :", numeric_end_time)

		fmt.Fprintln(w, "window_seconds     :", numeric_window_seconds)

		fmt.Fprintln(w, "url                :", token_params.url)

		fmt.Fprintln(w, "acl                :", token_params.acl)

		fmt.Fprintln(w, "key                :", token_params.key)

		fmt.Fprintln(w, "payload            :", token_params.payload)

		fmt.Fprintln(w, "algo               :", token_params.algo)

		fmt.Fprintln(w, "salt               :", token_params.salt)

		fmt.Fprintln(w, "session_id         :", token_params.session_id)

		fmt.Fprintln(w, "field_delimiter    :", token_params.field_delimiter)

		fmt.Fprintln(w, "acl_delimiter      :", token_params.acl_delimiter)

		fmt.Fprintln(w, "escape_early       :", token_params.escape_early)

		fmt.Fprintln(w, "escape_early_upper :", token_params.escape_early_upper)

		fmt.Fprintln(w, "verbose            :", token_params.verbose)

		fmt.Fprintln(w, "")

	}

	new_token := ""

	if token_params.ip != "" {

		new_token += "ip=" + escapeEarly(token_params, token_params.ip) + token_params.field_delimiter

	}


	if token_params.start_time != "" {

		new_token += "st=" + strconv.FormatInt(numeric_start_time, 10) + token_params.field_delimiter

	}


	new_token += "exp=" + strconv.FormatInt(numeric_end_time, 10) + token_params.field_delimiter

	if token_params.acl != "" {

		new_token += "acl=" + escapeEarly(token_params, token_params.acl) + token_params.field_delimiter

	}

	if token_params.session_id != "" {

		new_token += "id=" + escapeEarly(token_params, token_params.session_id) + token_params.field_delimiter

	}

	if token_params.payload != "" {

		new_token += "data=" + escapeEarly(token_params, token_params.payload) + token_params.field_delimiter

	}

	hash_source := new_token

	if token_params.url != "" && token_params.acl == "" {

		hash_source += "url=" + escapeEarly(token_params, token_params.url) + token_params.field_delimiter

	}

	if token_params.salt != "" {

		hash_source += "salt=" + token_params.salt + token_params.field_delimiter

	}


	var hmac_hash_source string

	if hash_source[len(hash_source)-1] == token_params.field_delimiter[0] {

		hmac_hash_source = hash_source[:len(hash_source)-1]

	} else {

		hmac_hash_source = hash_source[:]

	}


	key_bytes, key_err := hex.DecodeString(token_params.key)

	fmt.Println("hmac_hash_source is: ", hmac_hash_source)

	fmt.Printf("key_bytes is: %s\n", key_bytes)


	checkError(key_err, "Invalid hex encoded key")


	if strings.ToLower(token_params.algo) == "sha256" {

		h := hmac.New(sha256.New, key_bytes)

		_, write_err := h.Write([]byte(hmac_hash_source))

		checkError(write_err, "")

		new_token += "hmac=" + hex.EncodeToString(h.Sum(nil))

	} else if strings.ToLower(token_params.algo) == "sha1" {

		h := hmac.New(sha1.New, key_bytes)

		_, write_err := h.Write([]byte(hmac_hash_source))

		checkError(write_err, "")

		new_token += "hmac=" + hex.EncodeToString(h.Sum(nil))

	} else if strings.ToLower(token_params.algo) == "md5" {

		h := hmac.New(md5.New, key_bytes)

		_, write_err := h.Write([]byte(hmac_hash_source))

		checkError(write_err, "")

		new_token += "hmac=" + hex.EncodeToString(h.Sum(nil))

	} else {

		panic("Unknown algorithm.")

	}


	return token_params.token_name + "=" + new_token

}


func Token(rw http.ResponseWriter, req *http.Request) {

	defer func() {

		if r := recover(); r != nil {

			fmt.Fprintln(rw, r)

		}

	}()

	token_params := TokenParams{

		req.URL.Query().Get("token_type"),

		req.URL.Query().Get("token_name"),

		req.URL.Query().Get("ip"),

		req.URL.Query().Get("start_time"),

		req.URL.Query().Get("end_time"),

		req.URL.Query().Get("window"),

		req.URL.Query().Get("url"),

		req.URL.Query().Get("acl"),

		req.URL.Query().Get("key"),

		req.URL.Query().Get("payload"),

		req.URL.Query().Get("algo"),

		req.URL.Query().Get("salt"),

		req.URL.Query().Get("session_id"),

		req.URL.Query().Get("field_delimiter"),

		req.URL.Query().Get("acl_delimiter"),

		req.URL.Query().Get("escape_early") == "1",

		req.URL.Query().Get("escape_early_upper") == "1",

		req.URL.Query().Get("verbose") == "1",

	}

	// We need to delay the writing to the ResponseWriter so we can add headers.

	var b bytes.Buffer

	new_token := GenerateToken(&token_params, &b)

	rw.Header().Set("Server", "golang")

	rw.Header().Set(token_params.token_name, new_token)

	rw.Header().Set("Set-Cookie", new_token)

	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte(b.String() + new_token + "\n"))

}


var rest_help_string string = `

  acl string

   	Access control list delimited by ! [ie. /*]

  acl_delimiter string

   	Character used to delimit acl fields. [Default:!] (default "!")

  algo string

   	Algorithm to use to generate the token. (sha1, sha256, or md5) [Default:sha256] (default "sha256")

  end_time string

   	When does this token expire? --end_time overrides --window [Used for:URL or COOKIE]

  escape_early (0 or 1)

   	Causes strings to be url encoded before being used. (legacy 2.0 behavior)

  escape_early_upper (0 or 1)

   	Causes strings to be url encoded before being used. (legacy 2.0 behavior)

  field_delimiter string

   	Character used to delimit token body fields. [Default:~] (default "~")

  ip string

   	IP Address to restrict this token to.

  key string

   	Secret required to generate the token.

  payload string

   	Additional text added to the calculated digest.

  salt string

   	Additional data validated by the token but NOT included in the token body.

  session_id string

   	The session identifier for single use tokens or other advanced cases.

  start_time string

   	What is the start time. (Use now for the current time)

  token_name string

   	Parameter name for the new token. [Default:hdntl] (default "hdntl")

  token_type string

   	Select a preset: (optional) [2.0, 2.0.1 ,PV, Debug]

  url string

   	URL path. [Used for:URL]

  verbose (0 or 1)

   	Print extra details about the generation of the token.

  window string

    	How long is this token valid for?


curl "http://localhost:4000/token?token_name=hdnts&key=abc123&window=300&acl=/*"

hdnts=exp=1442142563~acl=/*~hmac=f626eb8e856f35b9cbb7ef6c9d9ea56a3c2993fe386f31fb12cb34bb083b069e


curl "http://localhost:4000/token?token_name=hdnts&key=abc123&window=300&acl=/*&verbose=1"


Akamai Token Generator v2.0.7

token_type         : 

token_name         : hdnts

algorithm          : sha256

start_time         : 0

end_time           : 1442142597

window_seconds     : 300

url                : 

acl                : /*

key                : abc123

payload            : 

algo               : sha256

salt               : 

session_id         : 

field_delimiter    : ~

acl_delimiter      : !

escape_early       : false

escape_early_upper : false

verbose            : true


hdnts=exp=1442142597~acl=/*~hmac=79742aa82fd2b9bcf3541316762530e6eedec673ac94d1429b315dff767003eb

`


func RestHelpString(rw http.ResponseWriter, req *http.Request) {

	defer func() {

		if r := recover(); r != nil {

			fmt.Fprintln(rw, r)

		}

	}()

	rw.Header().Set("Server", "golang")

	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte(rest_help_string))

}


func main() {

	var token_type = flag.String("token_type", "",

		"Select a preset: (optional) [2.0, 2.0.1 ,PV, Debug]")

	var token_name = flag.String("token_name", "hdntl",

		"Parameter name for the new token. [Default:hdntl]")

	var ip = flag.String("ip", "",

		"IP Address to restrict this token to.")

	var start_time = flag.String("start_time", "",

		"What is the start time. (Use now for the current time)")

	var end_time = flag.String("end_time", "",

		"When does this token expire? --end_time overrides --window [Used for:URL or COOKIE]")

	var window_seconds = flag.String("window", "",

		"How long is this token valid for?")

	var url = flag.String("url", "",

		"URL path. [Used for:URL]")

	var acl = flag.String("acl", "",

		"Access control list delimited by ! [ie. /*]")

	var key = flag.String("key", "",

		"Secret required to generate the token.")

	var payload = flag.String("payload", "",

		"Additional text added to the calculated digest.")

	var algo = flag.String("algo", "sha256",

		"Algorithm to use to generate the token. (sha1, sha256, or md5) [Default:sha256]")

	var salt = flag.String("salt", "",

		"Additional data validated by the token but NOT included in the token body.")

	var session_id = flag.String("session_id", "",

		"The session identifier for single use tokens or other advanced cases.")

	var field_delimiter = flag.String("field_delimiter", "~",

		"Character used to delimit token body fields. [Default:~]")

	var acl_delimiter = flag.String("acl_delimiter", "!",

		"Character used to delimit acl fields. [Default:!]")

	var escape_early = flag.Bool("escape_early", false,

		"Causes strings to be url encoded before being used. (legacy 2.0 behavior)")

	var escape_early_upper = flag.Bool("escape_early_upper", false,

		"Causes strings to be url encoded before being used. (legacy 2.0 behavior)")

	var verbose = flag.Bool("verbose", false,

		"Print extra details about the generation of the token.")

	var server = flag.Bool("server", false,

		"Start a web server on localhost for /help and /token")

	var server_port = flag.Int64("server_port", 4000,

		"Start a web server on localhost for /help and /token [Default:4000]")


	flag.Parse()

	if *server {

		server_port_str := strconv.FormatInt(*server_port, 10)

		fmt.Println("Listening on localhost:" + server_port_str)

		mux := http.NewServeMux()

		mux.HandleFunc("/", RestHelpString)

		mux.HandleFunc("/token", Token)

		log.Fatal(http.ListenAndServe("localhost:"+server_port_str, mux))

	} else {

		defer func() {

			if r := recover(); r != nil {

				fmt.Println(r)

			}

		}()

		token_params := TokenParams{

			*token_type,

			*token_name,

			*ip,

			*start_time,

			*end_time,

			*window_seconds,

			*url,

			*acl,

			*key,

			*payload,

			*algo,

			*salt,

			*session_id,

			*field_delimiter,

			*acl_delimiter,

			*escape_early,

			*escape_early_upper,

			*verbose,

		}

		new_token := GenerateToken(&token_params, os.Stdout)

		fmt.Println(new_token)

	}

}
