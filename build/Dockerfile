FROM golang:1.22.1

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.sum go.mod ./
COPY api/ api/
COPY internal/ internal/
COPY cmd/ cmd/
COPY ../configs/docker.yml configs/
COPY ../migrations/migrate.sql migrations/
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o rest_api ./cmd/rest_api/main.go

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 8080

