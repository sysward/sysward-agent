FROM golang
ENV DOCKER true
COPY . /sysward
WORKDIR /sysward
CMD ["make"]
