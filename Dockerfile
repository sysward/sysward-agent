FROM golang
ENV DOCKER true
ENV CGO_ENABLED 0
ENV SKIP_UPDATES true
COPY . /sysward
WORKDIR /sysward
CMD ["make"]
