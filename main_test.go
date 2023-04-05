package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/bobmaertz/canner/config"
)

func Test_contains(t *testing.T) {
	type args struct {
		reqd map[string]string
		hdrs http.Header
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		//make sure that the function returns true when the headers match
		{
			name: "headers match",
			args: args{
				reqd: map[string]string{
					"Content-Type": "application/json",
				},
				hdrs: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			want: true,
		},
		//make sure that the function returns false when the headers do not match
		{
			name: "headers do not match",
			args: args{
				reqd: map[string]string{
					"Content-Type": "application/json",
				},
				hdrs: http.Header{
					"Content-Type": []string{"application/xml"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := headersMatch(tt.args.reqd, tt.args.hdrs); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_waitFor(t *testing.T) {
	type args struct {
		latency config.LatencyConfig
		sleep   func(time.Duration)
	}
	tests := []struct {
		name string
		args args
	}{
		// make sure that the function sleeps for the correct amount of time
		{
			name: "simple latency",
			args: args{
				latency: config.LatencyConfig{
					Type:  "simple",
					Delay: 1 * time.Second,
				},
				sleep: func(d time.Duration) {

					if d != 1*time.Second {
						t.Errorf("waitFor() = %v, want %v", d, 1*time.Second)
					}
				},
			},
		},
		// make sure that the function sleeps for the correct amount of time
		{
			name: "random latency",
			args: args{
				latency: config.LatencyConfig{
					Type:  "random",
					Delay: 1 * time.Second,
				},
				sleep: func(d time.Duration) {
					if d > 1*time.Second {
						t.Errorf("waitFor() = %v, want %v", d, 1*time.Second)
					}
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			waitFor(tt.args.latency, tt.args.sleep)
		})
	}
}

func Test_bodyMatches(t *testing.T) {
	type args struct {
		reqBody      string
		incomingBody io.ReadCloser
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// make sure that the function returns true when the bodies match
		{
			name: "bodies match",
			args: args{
				reqBody: "test",
				incomingBody: io.NopCloser(
					bytes.NewBuffer([]byte("test")),
				),
			},
			want: true,
		},

		// make sure that the function returns false when the bodies do not match
		{
			name: "bodies do not match",
			args: args{
				reqBody: "test",
				incomingBody: io.NopCloser(
					bytes.NewBuffer([]byte("test2")),
				),
			},
			want: false,
		},

		// make sure that the function returns true when the bodies are empty
		{
			name: "bodies are empty",
			args: args{
				reqBody: "",
				incomingBody: io.NopCloser(
					bytes.NewBuffer([]byte("")),
				),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bodyMatches(tt.args.reqBody, tt.args.incomingBody); got != tt.want {
				t.Errorf("bodyMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createHandler(t *testing.T) {
	type args struct {
		matchers []Matcher
	}
	tests := []struct {
		name string
		args args
		want func(w http.ResponseWriter, incoming *http.Request)
	}{
		{
			name: "writes the correct response",
			args: args{
				matchers: []Matcher{
					{
						Request: config.Request{
							Method: "GET",
							Headers: map[string]string{
								"Content-Type": "application/json",
							},
						},
						Response: config.Response{
							StatusCode: http.StatusOK,
							Body:       "test",
						},
					},
				},
			},
			want: func(w http.ResponseWriter, incoming *http.Request) {
				w.WriteHeader(http.StatusOK)

				_, err := w.Write([]byte("test"))
				if err != nil {
					return
				}
			},
		},
		{
			name: "writes the correct response for no mocks found due to header mismatch",
			args: args{
				matchers: []Matcher{
					{
						Request: config.Request{
							Method: "GET",
							Headers: map[string]string{
								"Content-Type": "no match",
							},
						},
						Response: config.Response{
							StatusCode: http.StatusOK,
							Body:       "test",
						},
					},
				},
			},
			want: func(w http.ResponseWriter, incoming *http.Request) {
				w.WriteHeader(http.StatusNotFound)

				_, err := w.Write([]byte("mock not found"))
				if err != nil {
					return
				}
			},
		},
		{
			name: "writes the correct response for no mocks found due to header mismatch",
			args: args{
				matchers: []Matcher{
					{
						Request: config.Request{
							Method: "POST", //Not correct method
							Headers: map[string]string{
								"Content-Type": "application/json",
							},
						},
						Response: config.Response{
							StatusCode: http.StatusOK,
							Body:       "test",
						},
					},
				},
			},
			want: func(w http.ResponseWriter, incoming *http.Request) {
				w.WriteHeader(http.StatusNotFound)

				_, err := w.Write([]byte("mock not found"))
				if err != nil {
					return
				}
			},
		},
		{
			name: "writes the correct response for no mocks found due to body mismatch",
			args: args{
				matchers: []Matcher{
					{
						Request: config.Request{
							Method: "GET",
							Headers: map[string]string{
								"Content-Type": "application/json",
							},
							Body: "no match on the body",
						},
						Response: config.Response{
							Headers: map[string]string{
								"Content-Type": "application/json",
							},
							StatusCode: http.StatusOK,

							Body: "test response",
						},
					},
				},
			},
			want: func(w http.ResponseWriter, incoming *http.Request) {
				w.WriteHeader(http.StatusNotFound)

				_, err := w.Write([]byte("mock not found"))
				if err != nil {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createHandler(tt.args.matchers); !reflect.DeepEqual(got, tt.want) {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("Content-Type", "application/json")

				w1 := httptest.NewRecorder()
				w2 := httptest.NewRecorder()

				got(w1, req)
				tt.want(w2, req)
				if !reflect.DeepEqual(w1, w2) {
					t.Errorf("createHandler() = %v, want %v", w1, w2)
				}
			}
		})
	}
}

func Test_createMatchers(t *testing.T) {
	type args struct {
		c config.Configurations
	}
	tests := []struct {
		name string
		args args
		want map[string][]Matcher
	}{
		// TODO: Add test cases.
		// make sure that the function returns the correct matchers
		{
			name: "returns the correct matchers",
			args: args{
				c: config.Configurations{
					Matchers: []config.Matchers{
						{
							Request: config.Request{
								Path:   "/hello_world",
								Method: "GET",
								Headers: map[string]string{
									"Content-Type": "application/json",
								},
							},
							Response: config.Response{
								StatusCode: http.StatusOK,
								Body:       "test",
							},
						},
					},
				},
			},
			want: map[string][]Matcher{
				"/hello_world": {
					{
						Request: config.Request{
							Path:   "/hello_world",
							Method: "GET",
							Headers: map[string]string{
								"Content-Type": "application/json",
							},
						},
						Response: config.Response{
							StatusCode: http.StatusOK,
							Body:       "test",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createMatchers(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createMatchers() = %v, want %v", got, tt.want)
			}
		})
	}
}
