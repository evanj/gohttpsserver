# Go HTTPS Server and Proxy

Serves local files over HTTPS using a self-signed certificate. The certificate is generated when the server starts.

## Usage

1. `go run cmd/https_server.go`
2. `open https://localhost:8000/`


## Proxy

Proxies HTTPS requests to an HTTP server. Quick and dirty hack I use to test an HTTPS app locally. In production, we use nginx to decode HTTPS. This is a single binary to simulate that environment. It sends `X-Forwarded-For` and `X-Forwarded-Proto` headers, and passes through the `Host` header.

1. `go run cmd/https_proxy.go http://localhost:5000/`
2. `open https://localhost:8001/`
