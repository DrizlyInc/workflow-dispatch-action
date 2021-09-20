FROM golang:1.16-alpine AS builder

WORKDIR /build
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

#  Get dependencies before build
COPY action/go.mod action/go.sum ./
# RUN --mount=type=cache,target=$GOPATH/pkg/mod go mod download
RUN go mod download

# Build binary
COPY action/. .
RUN go build -o app

FROM builder AS linter
ENTRYPOINT GOFMT_OUTPUT="$(gofmt -d -e .)"; if [ -n "$GOFMT_OUTPUT" ]; then echo "${GOFMT_OUTPUT}"; exit 1 ; fi

FROM builder AS unit-tester
CMD go test -v ./...

FROM gcr.io/distroless/base:nonroot AS production
# set user to nonroot
USER nonroot
WORKDIR /
COPY --from=builder /build/app /app

ENTRYPOINT [ "/app" ]