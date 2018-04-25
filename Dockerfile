FROM debian:jessie

# Can't just apt-get install libpcap because this is the recommended version and ubuntu only had 1.5.3-2
RUN apt-get -y update && \
    apt-get install -y curl wget flex bison make build-essential && \
    curl -L https://github.com/Clever/gor/releases/download/v0.13.6/gor_0.13.6_x64.tar.gz | tar xvz -C /usr/local/bin/ && \
    chmod +x /usr/local/bin/gor && \
    wget http://www.tcpdump.org/release/libpcap-1.7.4.tar.gz && \
    tar xzf libpcap-1.7.4.tar.gz && \
    cd libpcap-1.7.4 && \
    ./configure && \
    make install && \
    cd .. && \
    rm -rf libpcap-1.7.4 && \
    apt-get install -y ca-certificates

COPY bin/sfncli /usr/bin/sfncli
COPY build/linux-amd64/http-science /usr/local/bin/http-science

CMD ["/usr/bin/sfncli", "--cmd", "/usr/local/bin/http-science", "--activityname", "${_DEPLOY_ENV}--${_APP_NAME}", "--region", "us-west-2", "--cloudwatchregion", "us-west-1", "--workername", "MAGIC_ECS_TASK_ARN"]
