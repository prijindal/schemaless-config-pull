FROM golang:1.24 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY cmd/ ./cmd/
COPY pkg/ ./cmd/

ARG GO_CMD TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /$GO_CMD cmd/$GO_CMD/main.go

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

ARG GO_CMD
COPY --from=build-stage /$GO_CMD /bin/main

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT ["/bin/main"]
