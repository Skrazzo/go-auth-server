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
- Docker (optional, for containerized deployment)

## Installation

### Option 1: Clone and Build

```bash
# Clone the repository
git clone https://github.com/yourusername/caddy-auth.git
cd caddy-auth

# Create and configure .env file
cat > .env << EOF
JWT_KEY=your_secure_jwt_secret_key
PORT=8080
COOKIE_NAME=caddy_auth_session
EOF

# Build the application
go build -o auth-server

# Run the server
./auth-server
```

### Option 2: Docker

```bash
# Clone the repository
git clone https://github.com/yourusername/caddy-auth.git
cd caddy-auth

# Create .env file (as above)

# Build and run with Docker
docker build -t caddy-auth .
docker run -p 8080:8080 --env-file .env caddy-auth
```

## Configuration

### Environment Variables

Create a `.env` file with the following variables:

```
JWT_KEY=your_secure_jwt_secret_key  # Secret key for signing JWT tokens
PORT=8080                           # Port for the auth server to listen on
COOKIE_NAME=caddy_auth_session      # Name of the authentication cookie
```

### User Management

Users are defined in `users/db.go`. For production use, you should modify this to use a proper database or authentication system.

Default user:
- Username: `username`
- Password: `plain-password`

## Caddy Configuration

Add the following to your Caddyfile to protect routes with authentication:

```
{
  # Global options
  debug
}

# Your auth service
auth.example.com {
  reverse_proxy localhost:8080
}

# Protected application
app.example.com {
  forward_auth auth.example.com/verify {
    uri /verify
    copy_headers X-Forwarded-Uri
  }
  
  reverse_proxy localhost:3000
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
2. Replace the basic user map with a proper database
3. Use a strong, randomly generated JWT key
4. Consider implementing rate limiting for login attempts

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.