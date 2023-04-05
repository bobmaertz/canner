package main

import (
	"net/http"
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
			if got := contains(tt.args.reqd, tt.args.hdrs); got != tt.want {
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
