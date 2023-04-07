package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bobmaertz/canner/config"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	seed = func() int64 {
		return time.Now().UnixNano()
	}
)

type Matcher struct {
	Request  config.Request
	Response config.Response
}

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
	// Set the file name of the configurations file
	viper.SetConfigName("config.yml")

	// Set the path to look for the configurations file
	viper.AddConfigPath("./conf")

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yml")

}

func main() {

	var c config.Configurations

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	rtr := mux.NewRouter()

	// handle http first
	// process all http requests and parse by url
	// we need to create a map with all the possible responses matched to each request.

	matchers := createMatchers(c)

	for path, matcher := range matchers {
		rtr.HandleFunc(path, createHandler(matcher))
	}

	http.Handle("/", rtr)

	log.Infof("Listening on %d\n", c.Server.Port)
	srv := &http.Server{
		Handler: rtr,
		Addr:    fmt.Sprintf("localhost:%d", c.Server.Port),

		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func createHandler(matchers []Matcher) func(w http.ResponseWriter, incoming *http.Request) {

	return func(w http.ResponseWriter, incoming *http.Request) {
		log.Debugf("Request Log: %v", incoming)

		var m *Matcher
		for _, r := range matchers {
			if !methodMatches(r.Request.Method, incoming.Method) {
				fmt.Println("method does not match")
				continue
			}

			if !headersMatch(r.Request.Headers, incoming.Header) {
				fmt.Println("header does not match")
				continue
			}

			if !bodyMatches(r.Request.Body, incoming.Body) {
				fmt.Println("body does not match")
				continue
			}

			//Set floating matcher var to the matcher we found
			m = &r
		}

		if m != nil {
			if m.Response.Latency != nil {
				waitFor(*m.Response.Latency, time.Sleep)
			}
			//Everything passed; return value
			for k, v := range m.Response.Headers {
				log.Debugf("key: %s, value: %s", k, v)
				w.Header().Add(k, v)
			}

			w.WriteHeader(m.Response.StatusCode)

			_, err := w.Write([]byte(m.Response.Body))
			if err != nil {
				return
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)

		_, err := w.Write([]byte("mock not found"))
		if err != nil {
			return
		}
	}
}

func waitFor(latency config.LatencyConfig, sleep func(time.Duration)) {
	switch latency.Type {
	case "random":
		rand.Seed(seed())
		max := latency.Delay.Nanoseconds()

		s := rand.Intn(int(max))
		sleep(time.Duration(s) * time.Nanosecond)

	case "simple":
		sleep(latency.Delay)
	default:
		//none
	}
}

func methodMatches(reqd string, incoming string) bool {
	if reqd == "" {
		reqd = "GET" //default to GET
	}

	if reqd == strings.ToUpper(incoming) {
		return true
	}
	return false
}

func headersMatch(reqd map[string]string, hdrs http.Header) bool {
	if len(reqd) == 0 {
		return true //no headers to match
	}
	for k, v := range reqd {
		val := hdrs.Get(k)
		if val == v {
			return true
		}
	}
	return false
}

func bodyMatches(reqBody string, incomingBody io.ReadCloser) bool {
	if reqBody == "" {
		return true //no body to match
	}

	body, err := io.ReadAll(incomingBody)
	if err != nil {
		log.Errorf("Error reading body: %v", err)
		return false
	}
	if string(body) == reqBody {
		return true
	}
	return false
}

func createMatchers(c config.Configurations) map[string][]Matcher {
	matchers := make(map[string][]Matcher)
	for _, m := range c.Matchers {
		mtchr := Matcher{
			Request:  m.Request,
			Response: m.Response,
		}
		matchers[m.Request.Path] = append(matchers[m.Request.Path], mtchr)
	}
	return matchers
}
