package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PrinceNorin/rburlshortener/transport"
)

var (
	fs     = flag.NewFlagSet("http", flag.ExitOnError)
	port   = fs.Int("port", 8080, "HTTP server port")
	logger = log.New(os.Stdout, "[HTTP] ", log.Ldate|log.Ltime)
)

func main() {
	h := transport.NewHTTPHandler()
	s := &http.Server{
		Handler:      h,
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("Listening to: http://127.0.0.1:%d", *port)
	if err := s.ListenAndServe(); err != nil {
		logger.Fatalf("Error: %v", err)
	}
}
