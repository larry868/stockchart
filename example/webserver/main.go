package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {

	const staticdir = "../webapp"

	fs := http.FileServer(http.Dir(staticdir))

	log.Print("Serving " + staticdir + " on http://localhost:5500")

	log.Fatal(http.ListenAndServe(":5500", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		// for dev, unset cache
		resp.Header().Add("Cache-Control", "no-cache")
		// apply a specific header for .wasm file
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(resp, req)
	})))

}
