# +-----------------------------------------------------------------------------
# | Pull vendor dependencies
# +-----------------------------------------------------------------------------
FROM golang:1.22.3-alpine as vendor 

RUN mkdir /app && mkdir /build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# +-----------------------------------------------------------------------------
# | Base layer for the hot-reloading development images
# +-----------------------------------------------------------------------------
FROM vendor as hot-reload

RUN go install github.com/cosmtrek/air@latest

CMD [ "air" ]

# +-----------------------------------------------------------------------------
# | Build the binaries
# +-----------------------------------------------------------------------------
FROM vendor AS builder

COPY . .
RUN go build -o /build/apiserver cmd/apiserver/main.go

# +-----------------------------------------------------------------------------
# | Create the final image
# +-----------------------------------------------------------------------------
FROM alpine:3.19
COPY ./static ./static
COPY --from=builder /build/* /app/main