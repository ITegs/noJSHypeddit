# +-----------------------------------------------------------------------------
# | Build the application
# +-----------------------------------------------------------------------------
FROM golang:1.22.3-alpine3.18 as builder

RUN mkdir /app && mkdir /build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /build/main main.go

# +-----------------------------------------------------------------------------
# | Create the final image
# +-----------------------------------------------------------------------------
FROM alpine:3.19
COPY ./static ./static
COPY --from=builder /build/main /app/main
CMD ["/app/main"]