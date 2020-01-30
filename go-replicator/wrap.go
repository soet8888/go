package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	ua "github.com/avct/uasurfer"
)

func jsonWrap(fn func(http.ResponseWriter, *http.Request) Return) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Cors(w, r)
		w.Header().Set("Content-Type", "application/json")

		var rt interface{}
		var enc *json.Encoder
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			rt = fn(w, r)
			enc = json.NewEncoder(w)
		} else {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()

			rt = fn(w, r)

			enc = json.NewEncoder(gz)
		}

		// Encode result
		if err := enc.Encode(&rt); nil != err {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"status":"Error","message":"%s"}`, err)
			fmt.Fprintf(w, `{ "Result"}`, err)
		}
	}
}
func authWrap(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := checkToken(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			rt := NewErrorStatus().SetMessage("Permission denied")
			// Encode result
			enc := json.NewEncoder(w)
			if err := enc.Encode(&rt); nil != err {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, `{"status":"Error","message":"%s"}`, err)
			}
		} else {
			fn(w, r)
		}

	}
}
func NoCache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Set our NoCache headers
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
func NoCacheFunc(h http.Handler) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Set our NoCache headers
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	}
	return fn
}

var epoch = time.Unix(0, 0).Format(time.RFC1123)
var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

func corsWrap(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Cors(w, r)
		fn(w, r)
	}
}

// Write CORS headers
func Cors(w http.ResponseWriter, r *http.Request) {
	var bwr ua.UserAgent
	ua.ParseUserAgent(r.Header.Get("User-Agent"), &bwr)
	if bwr.Browser.Name == ua.BrowserSafari {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Cache-Control,Content-Type,Expires,Last-Modified,Pragma,Token,Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Accept-Encoding,Accept-Language,Content-Type,Accept,User-Agent,X-Requested-With,Origin,Token,Browser,Referer")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
	}

}
