FROM golang:alpine as build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o video_service ./cmd/

# Build
FROM alpine
WORKDIR /app
COPY --from=build /app/video_service ./cmd/
CMD [ "./video_service" ]