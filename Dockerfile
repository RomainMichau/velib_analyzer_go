FROM golang:1.19

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY ./server/go.mod ./server/go.sum ./
RUN go mod download && go mod verify

WORKDIR /usr/src/app/server
COPY ./server .
RUN go build  -v -o /usr/local/bin/velib_analyzer ./.

ENTRYPOINT ["velib_analyzer"]