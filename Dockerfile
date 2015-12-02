FROM debian:jessie
COPY build/linux-amd64/http-science /usr/local/bin/http-science
CMD ["http-science"]
