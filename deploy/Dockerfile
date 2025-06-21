#build
FROM golang:alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN apk add upx && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o video_service ./cmd/ && upx --best --lzma video_service


#final-stage
FROM alpine
WORKDIR /app
COPY --from=build /app/video_service ./cmd/
CMD [ "./cmd/video_service" ]