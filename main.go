package main

import (
	"fmt"
	"github.com/bobmaertz/canner/config"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"math/rand"
	"net/http"
	"os"
	"time"
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
		return
	}
	err = viper.Unmarshal(&c)

	rtr := mux.NewRouter()

	// handle http first
	// process all http requests and parse by url
	// we need to create a map with all the possible responses matched to each request.
	matchers := make(map[string][]Matcher)
	for _, m := range c.Matchers {
		mtchr := Matcher{
			Request:  m.Request,
			Response: m.Response,
		}
		matchers[m.Request.Path] = append(matchers[m.Request.Path], mtchr)
	}

	for path, matcher := range matchers {
		rtr.HandleFunc(path, func(w http.ResponseWriter, incoming *http.Request) {
			log.Debugf("Request Log: %v", incoming)

			var m *Matcher
			for _, r := range matcher {
				if r.Request.Method != incoming.Method {
					continue
				}
				// Does each header match?
				if !contains(r.Request.Headers, incoming.Header) {
					continue
				}

				//TODO: does the body match?
				//Set floating matcher var to the matcher we found
				m = &r
			}

			if m != nil {
				if m.Response.Latency != nil {
					switch m.Response.Latency.Type {
					case "random":
						rand.Seed(seed())
						max := m.Response.Latency.Delay.Nanoseconds()

						s := rand.Intn(int(max))
						time.Sleep(time.Duration(s) * time.Nanosecond)

					case "simple":
						time.Sleep(m.Response.Latency.Delay)
					default:
						//none
					}
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
		})
	}
	http.Handle("/", rtr)

	log.Infof("Listening on %d\n", c.Server.Port)
	srv := &http.Server{
		Handler: rtr,
		Addr:    fmt.Sprintf("localhost:%d", c.Server.Port),
		//TODO: Add timeout options
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func contains(reqd map[string]string, hdrs http.Header) bool {
	for k, v := range reqd {
		val := hdrs.Get(k)
		if val == v {
			return true
		}
	}
	return false
}
