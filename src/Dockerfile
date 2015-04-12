FROM golang
RUN apt-get update -y --fix-missing
RUN apt-get install -y --fix-missing update-notifier
RUN apt-get install -y rsyslog
RUN apt-get install -y python-apt
RUN apt-get install -y tmux
# RUN touch /etc/crontab
# ENV DOCKER true
# ENV CGO_ENABLED 0
ENV SKIP_UPDATES true
COPY . /sysward
WORKDIR /sysward
RUN /etc/init.d/rsyslog start
RUN make deps
CMD ["/bin/bash"]
