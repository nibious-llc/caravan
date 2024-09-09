############################
# STEP 1 build executable binary
############################
FROM golang@sha256:f475434ea2047a83e9ba02a1da8efc250fa6b2ed0e9e8e4eb8c5322ea6997795 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001
# See https://stackoverflow.com/a/55757473/12429735RUN 
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

WORKDIR /go/app/
COPY go.mod .
COPY go.sum .
# Fetch dependencies.
# Using go mod with go 1.11
RUN go mod download
RUN go mod verify

COPY . .
# Build the binary
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/app/bin/caravan-server cmd/server/main.go

############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# Copy our static executable
COPY --from=builder /go/app/bin/caravan-server /go/bin/caravan-server
# Use an unprivileged user.
USER appuser:appuser
# Run the hello binary.
ENTRYPOINT ["/go/bin/caravan-server"]
