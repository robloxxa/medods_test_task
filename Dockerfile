FROM golang:1.21.0-alpine3.18 as build
WORKDIR /app
COPY go.mod go.sum ./
COPY *.go .
COPY .env .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /medods-test-task

FROM alpine:3.18
COPY --from=build /medods-test-task /medods-test-task

ENTRYPOINT [ "/medods-test-task" ]