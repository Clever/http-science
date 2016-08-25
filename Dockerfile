FROM debian:jessie

# Can't just apt-get install libpcap because this is the recommended version and ubuntu only had 1.5.3-2
RUN apt-get -y update && apt-get install -y curl wget flex bison make build-essential && \
  curl -L https://github.com/Clever/gor/releases/download/v0.13.2/gor_0.13.2_x64.tar.gz | tar xvz -C /usr/local/bin/ && chmod +x /usr/local/bin/gor && \
  wget http://www.tcpdump.org/release/libpcap-1.7.4.tar.gz && tar xzf libpcap-1.7.4.tar.gz && cd libpcap-1.7.4 && \
    ./configure && make install && cd .. && rm -rf libpcap-1.7.4 && \
  curl -L https://github.com/Clever/gearcmd/releases/download/v0.8.0/gearcmd-v0.8.0-linux-amd64.tar.gz | \
    tar xz -C /usr/local/bin --strip-components 1

COPY build/linux-amd64/http-science /usr/local/bin/http-science

CMD ["gearcmd", \
    "--name", "http-science", \
    "--cmd", "/usr/local/bin/http-science", \
    "--parseargs=false", \
    "pass-sigterm"]
