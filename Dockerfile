FROM golang:1.18.6-alpine3.16 AS build
LABEL stage=intermediate
# mark it as working directory
WORKDIR /go/src/github.com/adibaulia/anon-twt
# Copy everything in this directory to Workdir
RUN apk update && apk add --no-cache git
COPY . .
RUN cp -R config/ /go/bin/config/
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin cmd/main.go

FROM alpine:3.10
RUN apk --no-cache add ca-certificates tzdata
RUN cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime 
RUN echo "Asia/Jakarta" >  /etc/timezone && date
RUN apk del tzdata

ENV TZ=Asia/Jakarta

WORKDIR /usr/bin
COPY entry.sh /usr/bin/entry.sh
COPY keys.json /usr/bin/keys.json
COPY --from=build /go/bin .
RUN chmod +x entry.sh

EXPOSE 80
ENTRYPOINT ["/usr/bin/entry.sh"]