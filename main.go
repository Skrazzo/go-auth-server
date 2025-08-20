package main

import (
	"caddy-auth/cache"
	"caddy-auth/routes"
	"embed"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
)

// Init gets automatically called
func init() {
	// Check for environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("[ERROR] while loading .env file: \"%v\"", err)
	}

	keysToCheck := []string{"JWT_KEY", "PORT", "COOKIE_NAME", "USERNAME", "PASSWORD", "EXPIRE_IN"}
	for _, key := range keysToCheck {
		if os.Getenv(key) == "" {
			log.Fatalf("[ERROR] Please set your %s in environment variable", key)
		}
	}

	// Load cache
	cache.Load()
}

// For embeded templates
//
//go:embed templates/*.html
var templatesFs embed.FS

func main() {
	// Close cache on exit
	defer cache.Close()
	// Load templates
	loginTemplate := template.Must(template.ParseFS(templatesFs, "templates/login.html"))

	// Load routes class
	r := routes.Routes{}
	r.Login.Tmpl = *loginTemplate

	// Start server listen
	http.HandleFunc("/login", r.LoginHandler)
	http.HandleFunc("/verify", r.VerifyHandler)
	http.HandleFunc("/", r.AliveHandler)

	log.Println("Server started on port :" + os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
