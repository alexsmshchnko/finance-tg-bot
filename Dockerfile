FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg
COPY config ./config
COPY internal ./internal

# Build the Go binary
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o finance-tg-bot ./cmd/*.go
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o finance-tg-bot ./cmd/*.go

# Create a minimal production image
FROM alpine:3.18

# It's essential to regularly update the packages within the image to include security patches
RUN apk update && apk upgrade

# Reduce image size
RUN rm -rf /var/cache/apk/* && \
    rm -rf /tmp/*

# Avoid running code as a root user
RUN adduser -D appuser
USER appuser

# Set the working directory inside the container
WORKDIR /app

# Copy only the necessary files from the builder stage
COPY --from=builder /app/config/config.local.hcl config.hcl 
COPY --from=builder /app/finance-tg-bot .

EXPOSE 8080

CMD ["./finance-tg-bot"]