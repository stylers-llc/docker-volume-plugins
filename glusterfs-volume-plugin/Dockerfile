FROM oraclelinux:7-slim
RUN yum install -q -y oracle-gluster-release-el7 && \
    yum install -q -y git glusterfs glusterfs-fuse attr && \
    curl --silent -L https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz | tar -C /usr/local -zxf -
RUN mkdir -p /root/go/src/glusterfs
COPY mounted-volume /root/go/src/mounted-volume
COPY glusterfs-volume-plugin /root/go/src/glusterfs-volume-plugin

RUN export GOPATH=/root/go && \
    /usr/local/go/bin/go get -u github.com/docker/go-plugins-helpers/volume && \
    /usr/local/go/bin/go get -u github.com/boltdb/bolt && \
    /usr/local/go/bin/go get -u github.com/dchest/uniuri && \
    cd /root/go/src/glusterfs-volume-plugin && /usr/local/go/bin/go install --ldflags '-extldflags "-static"'
RUN mv /root/go/bin/glusterfs-volume-plugin /
RUN rm -rf /root/go /usr/local/go && \
    yum remove -q -y git && \
    yum autoremove -q -y && \
    yum clean all && \
    rm -rf /var/cache/yum /var/log/anaconda /var/cache/yum /etc/mtab && \
    rm /var/log/lastlog /var/log/tallylog
