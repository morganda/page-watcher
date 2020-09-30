# Largely based on: https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324

############################
# STEP 1 build executable binary
############################
FROM golang@sha256:0978cc067eb3f53901c00b70a024f182baa371bdfe7f35f3d64e56cab2471c4d as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apt update && apt install git ca-certificates && update-ca-certificates

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

WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .


# Fetch dependencies.

# Using go mod with go 1.11
RUN go mod download
RUN go mod verify

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/page-watcher

############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static executable
COPY --from=builder /go/bin/page-watcher /go/bin/page-watcher

# Use an unprivileged user.
USER appuser:appuser

# Run the binary.
ENTRYPOINT ["/go/bin/page-watcher"]