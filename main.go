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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)


type Matcher struct {
	Request  config.Request
	Response config.Response
}

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

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

	// Set default values for timeouts if not set by configs
	readTimeout := 15 * time.Second
	if c.Server.ReadTimeout != 0 {
		readTimeout = c.Server.ReadTimeout
	}

	writeTimeout := 15 * time.Second
	if c.Server.WriteTimeout != 0 {
		writeTimeout = c.Server.WriteTimeout
	}

	idleTimeout := 60 * time.Second
	if c.Server.IdleTimeout != 0 {
		idleTimeout = c.Server.IdleTimeout
	}

	readHeaderTimeout := 15 * time.Second
	if c.Server.ReadHeaderTimeout != 0 {
		readHeaderTimeout = c.Server.ReadHeaderTimeout
	}

	log.Infof("Listening on %d\n", c.Server.Port)
	srv := &http.Server{
		Handler: rtr,
		Addr:    fmt.Sprintf("localhost:%d", c.Server.Port),

		WriteTimeout:      writeTimeout,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}

func createHandler(matchers []Matcher) func(w http.ResponseWriter, incoming *http.Request) {

	return func(w http.ResponseWriter, incoming *http.Request) {
		trace := uuid.New().String()
		log.Infof("Request [id: %s]: %v", trace, incoming)

		// Read body once before matching to avoid consumption issues
		bodyBytes, err := io.ReadAll(incoming.Body)
		if err != nil {
			log.Errorf("error reading request body: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bodyString := string(bodyBytes)

		var m *Matcher
		for _, r := range matchers {
			if !methodMatches(r.Request.Method, incoming.Method) {
				log.Error("method does not match")
				continue
			}

			if !headersMatch(r.Request.Headers, incoming.Header) {
				log.Error("header does not match")
				continue
			}

			if !bodyMatches(r.Request.Body, bodyString) {
				log.Error("body does not match")
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
				w.Header().Add(k, v)
			}

			w.WriteHeader(m.Response.StatusCode)

			_, err := w.Write([]byte(m.Response.Body))
			if err != nil {
				log.Errorf("error writing response body: %v", err)
				return
			}
			log.Infof("Response [id: %s]: %v", trace, m.Response)
			return
		}
		w.WriteHeader(http.StatusNotFound)

		_, err = w.Write([]byte("mock not found"))
		if err != nil {
			log.Errorf("error writing not found response: %v", err)
			return
		}
	}
}

func waitFor(latency config.LatencyConfig, sleep func(time.Duration)) {
	switch latency.Type {
	case "random":
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

	if strings.ToUpper(reqd) == strings.ToUpper(incoming) {
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
		if val != v {
			return false
		}
	}
	return true
}

func bodyMatches(reqBody string, incomingBody string) bool {
	if reqBody == "" {
		return true //no body to match
	}

	if incomingBody == reqBody {
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
