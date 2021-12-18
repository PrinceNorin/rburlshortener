package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PrinceNorin/rburlshortener/service"
	"github.com/PrinceNorin/rburlshortener/transport"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	fs     = flag.NewFlagSet("http", flag.ExitOnError)
	port   = fs.Int("port", 8080, "HTTP server port")
	logger = log.New(os.Stdout, "[HTTP] ", log.Ldate|log.Ltime)
)

func main() {
	// connect to sqlite database
	db, err := gorm.Open(sqlite.Open("database.sqlite"), &gorm.Config{})
	checkError(err)
	// save created_at as UTC
	db.NowFunc = func() time.Time {
		return time.Now().UTC()
	}
	// create short_urls table. ideally we want to integrate
	// with some sort of migration management tool
	// but we skip it in this example for simplicity
	err = initSchema(db)
	checkError(err)

	// Load required environment variables
	host := env("SERVER_HOST")
	adminToken := env("ADMIN_TOKEN")
	// comma separated pattern of blacklist
	// normally should have api to manage blacklist
	blacklistPatterns := loadBlacklist()

	svc := service.NewURLShortener(service.NewURLShortenerRepository(db))
	svc, err = service.WithBlacklist(svc, blacklistPatterns)
	checkError(err)

	h := transport.NewHTTPHandler(transport.HTTPConfig{
		Service:    svc,
		ServerHost: host,
		AdminToken: adminToken,
	})
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

func initSchema(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&service.ShortURL{})
	return
}

func checkError(err error) {
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}
}

func env(key string) string {
	val := os.Getenv(key)
	if val == "" {
		checkError(fmt.Errorf("missing env [%s]", key))
	}
	return val
}

func loadBlacklist() []string {
	var patterns []string
	for _, pattern := range strings.Split(os.Getenv("BLACKLIST"), ",") {
		patterns = append(patterns, strings.TrimSpace(pattern))
	}
	return patterns
}
