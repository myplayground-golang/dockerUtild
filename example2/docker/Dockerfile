FROM debian:buster-20220125-slim
LABEL author="SH team"

COPY ./opt /opt

RUN echo 'start' && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        vim \
    && \
    mv /opt/.bashrc /root/.bashrc && \
    # entrypoint_vain.sh for not starting any service
    mv /opt/entrypoint_vain.sh /usr/local/bin && \
    chmod 777 /usr/local/bin/entrypoint_vain.sh && \
    ln -s /usr/local/bin/entrypoint_vain.sh /docker-entrypoint-vain.sh && \
    # create folder for lp3 workspace
    mkdir -p /lp3/workspace/ && \
    mkdir -p /lp3/tools/ && \
    # clean /opt
    apt -y install fontconfig && \
    fc-cache --force


ADD ./my_jdk/my_jdk.tar.gz /usr/local

COPY ./tomcat-sfs /lp3/workspace/tomcat-sfs
COPY ./showVersion.sh /bin/showVersion.sh
COPY ./showVersion.json /lp3/tools/showVersion.json

ENV JAVA_VERSION jvm_8
ENV JAVA_HOME=/usr/local/jvm_8
ENV CLASSPATH=$JAVA_HOME/bin
ENV PATH=.:$PATH:$JAVA_HOME/bin


# setup entrypoint
ENTRYPOINT ["/docker-entrypoint-vain.sh"]