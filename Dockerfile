FROM google/debian:wheezy

# Need curl for pprof...
RUN apt-get update
RUN apt-get -y install curl
COPY build/linux-amd64/http-science /usr/local/bin/http-science

CMD ["http-science"]
