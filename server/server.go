package main

import (
	"crypto/tls"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

//go:embed public
var f embed.FS

const (
	DIRECTORY_TO_SERVE    = "public" // NOTE: this should always equal the embedded directory
	CERTIFICATE_CACHE_DIR = "."
	DEFAULT_HTTP_PORT     = 80
	DEFAULT_HTTPS_PORT    = 443
)

func main() {
	httpsDomain := flag.String("https-domain", "", "The domain to use for HTTPS. HTTPS will be automatically used if this is set to a non-empty value.")
	port := flag.Int("port", 0, "Port to use. If the value is set to zero (the default value), 443 will be used for HTTPS and 80 for HTTP.")
	flag.Parse()

	https := httpsDomain != nil && *httpsDomain != ""

	mux := getMux()

	if https {
		if port != nil && *port == 0 {
			*port = DEFAULT_HTTPS_PORT
		}

		// Create an autocert manager.
		manager := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(*httpsDomain),
			Cache:      autocert.DirCache(CERTIFICATE_CACHE_DIR),
		}

		// Create a TLS config using the autocert manager.
		tlsConfig := &tls.Config{
			GetCertificate: manager.GetCertificate,
		}

		// Create an HTTP server with TLS.
		server := &http.Server{
			Addr:      fmt.Sprintf(":%d", *port),
			TLSConfig: tlsConfig,
			Handler:   mux,
		}

		// Redirect HTTP to HTTPS.
		go func() {
			http.ListenAndServe(":http", manager.HTTPHandler(nil))
		}()

		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		if port != nil && *port == 0 {
			*port = DEFAULT_HTTP_PORT
		}

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
	}
}

func getMux() *http.ServeMux {
	mux := http.NewServeMux()

	serve := func(mux *http.ServeMux, url string, path string) {
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != url {
				http.NotFound(w, r)
			} else {
				http.ServeFileFS(w, r, f, path)
			}
		})
	}

	fs.WalkDir(f, DIRECTORY_TO_SERVE, func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() {
			url := strings.TrimPrefix(path, DIRECTORY_TO_SERVE)

			if strings.HasSuffix(url, "index.html") {
				serve(mux, strings.TrimSuffix(url, "index.html"), path)
			} else if strings.HasSuffix(url, ".html") {
				serve(mux, strings.TrimSuffix(url, ".html"), path)
			} else {
				serve(mux, url, path)
			}
		}
		return nil
	})

	return mux
}
