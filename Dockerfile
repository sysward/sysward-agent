FROM golang
ENV DOCKER true
ENV CGO_ENABLED 0
COPY . /sysward
WORKDIR /sysward
CMD ["make"]
