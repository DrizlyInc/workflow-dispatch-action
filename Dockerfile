FROM golang:1.16-alpine AS builder

WORKDIR /build
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

#  Get dependencies before build
COPY go.mod go.sum ./
# RUN --mount=type=cache,target=$GOPATH/pkg/mod go mod download
RUN go mod download

# Build binary
COPY . .
RUN go build -o app

FROM gcr.io/distroless/base:nonroot AS production
# set user to nonroot
USER nonroot
WORKDIR /
COPY --from=builder /build/app /app

ENTRYPOINT [ "/app" ]