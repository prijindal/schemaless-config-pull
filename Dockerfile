FROM golang:1.24 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY cmd/ ./cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -o /config_pull cmd/config_pull/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /management_server cmd/management_server/main.go

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /config_pull /bin/config_pull
COPY --from=build-stage /management_server /bin/management_server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["management_server"]