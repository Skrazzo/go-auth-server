package routes

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"text/template"
	"time"

	jwtUtils "caddy-auth/jwt"

	"github.com/golang-jwt/jwt/v5"
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

		expireInDays, err := strconv.Atoi(os.Getenv("EXPIRE_IN"))
		if err != nil {
			log.Panicln("[ERROR] Converting EXPIRE_IN to int", err)
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user":   user,
			"expire": time.Now().Add(time.Hour * time.Duration(24*expireInDays)).Unix(),
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
			Expires:  time.Now().Add(time.Duration(24*expireInDays) * time.Hour), // Cookie valid for 24 hours
		}

		http.SetCookie(w, &cookie)

		// The user will then naturally try to access the protected app again.
		http.Redirect(w, req, redirectURI, http.StatusFound)
		return
	}

	r.Login.Tmpl.Execute(w, LoginPage{RedirectURI: redirectURI})
}

func (r *Routes) RedirectToLogin(w http.ResponseWriter, req *http.Request, originalUrl string) {
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
		r.RedirectToLogin(w, req, originalURI)
		return
	}

	// Add parsing here
	claims, err := jwtUtils.CachedParse(tokenStr.Value)
	if err != nil {
		// Error while parsing JWT
		log.Println("[ERROR] Parsing JWT /verify ", err)
		r.RedirectToLogin(w, req, originalURI)
		return
	}

	// Get data from jwt claims
	expire, ok := claims["expire"].(float64)
	if !ok {
		log.Println("[ERROR] Getting expire ", err)
		r.RedirectToLogin(w, req, originalURI)
		return
	}

	// Check if token is epxired, delete if it is, and Redirect to login
	if expire < float64(time.Now().Unix()) {
		log.Println("[ERROR] Cookie is expired")
		r.RedirectToLogin(w, req, originalURI)
		return
	}

	// Pass back to caddy
	// w.Header().Set("X-Authenticated-User", user)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authenticated"))
}

func (r *Routes) AliveHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world"))
}
