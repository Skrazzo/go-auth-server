package routes

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginPage struct {
	RedirectURI string
}

type Routes struct {
	Login struct {
		Tmpl template.Template
	}
}

func (r *Routes) LoginHandler(w http.ResponseWriter, req *http.Request) {
	// Extract redirect_uri from the URL
	redirectURI := req.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = "/" // Default redirect after login
	}

	if req.Method == "POST" {
		user := req.FormValue("username")
		pass := req.FormValue("password")
		redirectURI := req.FormValue("redirect_uri")

		if redirectURI == "" {
			redirectURI = "/" // Default redirect after login
		}

		// Incorrect username
		if os.Getenv("USERNAME") != user {
			r.Login.Tmpl.Execute(w, LoginPage{RedirectURI: redirectURI})
			return
		}

		// Incorrect password
		if os.Getenv("PASSWORD") != pass {
			r.Login.Tmpl.Execute(w, LoginPage{RedirectURI: redirectURI})
			return
		}

		// Generate hash
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			log.Panicln("[ERROR] hashing password", err)
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user":   user,
			"hash":   string(hash),
			"expire": time.Now().Add(time.Hour * 24).Unix(),
		})

		// Sign token
		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
		if err != nil {
			log.Panicln("[ERROR] Signing JWT token", err)

		}

		// Create and set cookie
		// 3600 -> seconds in 1 hour
		// Successful login. Set the authentication cookie.
		// httponly: prevents client-side script access to the cookie.
		// secure: ensures cookie is only sent over HTTPS (disable for http://localhost if needed for dev)
		// samesite: helps mitigate CSRF attacks.
		// path: applies cookie to the entire domain.
		// Expires: IMPORTANT for real-world scenarios, sets how long the cookie is valid.
		// The value here is just "logged_in_USERNAME" for simplicity, in a real app it would be a secure, random session ID.
		cookie := http.Cookie{
			Name:     os.Getenv("COOKIE_NAME"),
			Value:    tokenString,
			Path:     "/",
			HttpOnly: true,
			Secure:   true, // Set to true for HTTPS in production!
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(24 * time.Hour), // Cookie valid for 24 hours
		}

		http.SetCookie(w, &cookie)

		// The user will then naturally try to access the protected app again.
		http.Redirect(w, req, redirectURI, http.StatusFound)
		return
	}

	r.Login.Tmpl.Execute(w, LoginPage{RedirectURI: redirectURI})
}

func (r *Routes) redirectToLogin(w http.ResponseWriter, req *http.Request, originalUrl string) {
	// Encode the original URI to be safely included as a query parameter.
	loginRedirectURL := "/login?redirect_uri=" + url.QueryEscape(originalUrl)
	log.Printf("Redirecting unauthenticated user to %s (original URI: %s)", loginRedirectURL, originalUrl)
	http.Redirect(w, req, loginRedirectURL, http.StatusFound)
}

func (r *Routes) VerifyHandler(w http.ResponseWriter, req *http.Request) {
	// Get the original URI from Caddy's X-Original-URI header
	originalURI := req.Header.Get("X-Forwarded-Uri")
	if originalURI == "" {
		// Fallback or error handling if header is missing
		originalURI = "/" // Default to home page
	}

	// Extract token from cookie
	tokenStr, err := req.Cookie(os.Getenv("COOKIE_NAME"))
	if err != nil {
		r.redirectToLogin(w, req, originalURI)
		return
	}

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenStr.Value, func(token *jwt.Token) (interface{}, error) {
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_KEY")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		// Error while parsing JWT
		log.Println("[ERROR] Parsing JWT /verify ", err)
		r.redirectToLogin(w, req, originalURI)
		return
	}

	// Get the data from JWT, and verify hash with user
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[ERROR] Getting claims ", err)
		r.redirectToLogin(w, req, originalURI)
		return
	}

	// Compare user hash, and check if correct username is provided
	user, ok := claims["user"].(string)
	if !ok {
		log.Println("[ERROR] Getting user ", err)
		r.redirectToLogin(w, req, originalURI)
		return
	}

	hash, ok := claims["hash"].(string)
	if !ok {
		log.Println("[ERROR] Getting hash ", err)
		r.redirectToLogin(w, req, originalURI)
		return
	}

	expire, ok := claims["expire"].(float64)
	if !ok {
		log.Println("[ERROR] Getting expire ", err)
		r.redirectToLogin(w, req, originalURI)
		return
	}

	userPass := os.Getenv("PASSWORD")

	// Compare user hash, and password in db
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(userPass))
	if err != nil {
		log.Println("[ERROR] Passwords dont match ", err)
		r.redirectToLogin(w, req, originalURI)
		return
	}

	// Check if token is epxired, delete if it is, and Redirect to login
	if expire < float64(time.Now().Unix()) {
		log.Println("[ERROR] Cookie is expired")
		r.redirectToLogin(w, req, originalURI)
		return
	}

	// Pass back to caddy
	w.Header().Set("X-Authenticated-User", user)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authenticated"))
}

func (r *Routes) AliveHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world"))
}
