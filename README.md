# go-url-shortener-cicd-aws

A production-style URL Shortener written in Go, containerized with Docker, and deployed to AWS EC2 via a fully automated GitHub Actions CI/CD pipeline with integrated Trivy security scanning.

---

## Architecture

```
GitHub Push → GitHub Actions CI/CD Pipeline
                    │
                    ├── 1. Build & Test (go test)
                    ├── 2. Docker Build + Trivy Security Scan
                    ├── 3. Push to Docker Hub
                    └── 4. SSH Deploy to AWS EC2
```

---

## Tech Stack

| Layer | Tool |
|---|---|
| Language | Go 1.21 |
| Containerization | Docker (multi-stage build) |
| CI/CD | GitHub Actions |
| Container Registry | Docker Hub |
| Security Scanning | Trivy (Aqua Security) |
| Cloud | AWS EC2 (Amazon Linux / Ubuntu) |

---

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `POST` | `/shorten` | Shorten a URL |
| `GET` | `/:code` | Redirect to original URL |

### Example Usage

**Shorten a URL:**
```bash
curl -X POST http://<your-ec2-ip>/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/Amarachi-Ezeonyekwere"}'
```

**Response:**
```json
{
  "short_code": "aB3xZ1",
  "short_url": "http://<your-ec2-ip>/aB3xZ1",
  "original_url": "https://github.com/Amarachi-Ezeonyekwere"
}
```

**Redirect:**
```bash
curl -L http://<your-ec2-ip>/aB3xZ1
# → redirects to https://github.com/Amarachi-Ezeonyekwere
```

---

## CI/CD Pipeline

The GitHub Actions workflow (`.github/workflows/ci-cd.yml`) runs on every push to `main`:

1. **Build & Test** — compiles the Go binary and runs all unit tests with coverage
2. **Docker Build & Trivy Scan** — builds the Docker image and scans for CRITICAL/HIGH CVEs; pipeline fails if vulnerabilities are found
3. **Push to Docker Hub** — tags and pushes the image with both `latest` and the commit SHA
4. **Deploy to EC2** — SSHes into the EC2 instance, pulls the new image, and restarts the container with zero-downtime replacement

---

## Docker Image

Multi-stage build using `golang:1.21-alpine` → `scratch` final image.

- Final image size: **< 10MB**
- No OS, no shell, minimal attack surface

```bash
docker pull <your-dockerhub-username>/go-url-shortener-cicd-aws:latest

docker run -d -p 8080:8080 \
  -e APP_HOST=http://localhost:8080 \
  <your-dockerhub-username>/go-url-shortener-cicd-aws:latest
```

---

## GitHub Actions Secrets Required

Set these in your repository under **Settings → Secrets and variables → Actions**:

| Secret | Description |
|---|---|
| `DOCKERHUB_USERNAME` | Your Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token (not your password) |
| `EC2_HOST` | Public IP or DNS of your EC2 instance |
| `EC2_USER` | SSH username (`ubuntu`, `ec2-user`, etc.) |
| `EC2_SSH_KEY` | Private SSH key for EC2 access (PEM contents) |

---

## Running Locally

```bash
# Clone the repo
git clone https://github.com/Amarachi-Ezeonyekwere/go-url-shortener-cicd-aws.git
cd go-url-shortener-cicd-aws

# Run tests
go test ./... -v

# Run the app
go run main.go

# Or with Docker
docker build -t url-shortener .
docker run -p 8080:8080 url-shortener
```

---

## Project Structure

```
go-url-shortener-cicd-aws/
├── main.go                          # Application entrypoint + handlers
├── main_test.go                     # Unit tests
├── go.mod                           # Go module definition
├── Dockerfile                       # Multi-stage Docker build
├── .gitignore
└── .github/
    └── workflows/
        └── ci-cd.yml                # GitHub Actions pipeline
```

---

*Built as part of a hands-on DevOps portfolio — demonstrating end-to-end CI/CD, containerization, cloud deployment, and security scanning fundamentals.*