#build
FROM golang:alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN apk add upx && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o video_service ./cmd/ && upx --best --lzma video_service


#final-stage
FROM alpine
WORKDIR /app
RUN wget -q -O - https://github.com/zaac04/lockenv/releases/download/v2.1.2/install.sh | sh -s v2.1.2 lockenv-linux-server-amd64 
COPY --from=build /app/video_service ./cmd/
CMD [ "./cmd/video_service" ]