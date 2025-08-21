# Caddy Forward Auth

A lightweight authentication service for Caddy's forward_auth middleware. This project provides a unified authentication layer for self-hosted applications with minimal resource usage.

## Features

- Simple JWT-based authentication
- Login page with customizable styling
- Verification endpoint for Caddy's forward_auth
- Low memory footprint
- Easy to deploy and configure

## Prerequisites

- Go 1.16+ (for building)
- Caddy v2

## Installation

```bash
# Clone the repository
git clone https://github.com/Skrazzo/go-auth-server
# Install packages
go mod tidy
```

### Build the application

```sh
go build -o auth-server
```

### Run the server

```sh
# With go command
go run main.go

# Or built first and then, make app executable
chmod +x ./auth-server
# Run the server
./auth-server
```

## Configuration

### Environment Variables

Modify `.env` file with the following variables:

```
JWT_KEY=your_secure_jwt_secret_key  # Secret key for signing JWT tokens
PORT=8080                           # Port for the auth server to listen on
COOKIE_NAME=caddy_auth_session      # Name of the authentication cookie
USERNAME=                           # Username for authentication
PASSWORD=                           # Password for authentication
EXPIRE_IN=7                         # Expire in days (after it, user will be asked to log in)
```

### User Management

Users are defined in `users/db.go`.

Default user:

- Username: `username`
- Password: `plain-password`

## Caddy Configuration

Add the following to your Caddyfile to protect routes with authentication:

```
# Protected application
app.example.com {
  # Login route for auth server (Needs to be unprotected)
  handle /login* {
    reverse_proxy localhost:8080 # proxy to go-auth server
  }

  handle /* {
    # Forward auth will check if user is authenticated with /verify url
    forward_auth localhost:8080 {
      uri /verify
    }

    # Your protected application
    reverse_proxy localhost:19999
  }
}
```

## How It Works

1. When a user accesses a protected application:

    - Caddy forwards the request to the `/verify` endpoint
    - If authenticated, the request proceeds to the application
    - If not authenticated, the user is redirected to the login page

2. Authentication flow:

    - User submits login credentials
    - Server validates credentials and issues a JWT token
    - Token is stored in a secure cookie
    - User is redirected back to the original URL

3. Verification process:
    - The `/verify` endpoint checks for a valid cookie
    - If valid, it returns a 200 status code and sets an `X-Authenticated-User` header
    - If invalid, it redirects to the login page

## Endpoints

- `/login` - Login page and authentication endpoint
- `/verify` - Verification endpoint for Caddy's forward_auth
- `/` - Health check endpoint

## Building for Production

A build script is included to compile for Linux:

```bash
chmod +x build-server.sh
./build-server.sh
```

## Security Considerations

For production deployment:

1. Set `Secure: true` in the cookie configuration (routes.go) when using HTTPS
2. Use a strong, randomly generated JWT key

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
