FROM google/debian:wheezy

COPY build/linux-amd64/http-science /usr/local/bin/http-science

CMD ["http-science"]
