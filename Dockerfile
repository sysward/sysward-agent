FROM golang
ENV DOCKER true
ENV CGO_ENABLED 0
ENV SKIP_UPDATES true
WORKDIR /go/src/bitbucket.org/sysward/sysward-agent
ADD go.mod .
RUN go mod tidy
ADD . .
CMD ["make"]
