FROM golang:1.15.2 as build

RUN apt-get update -y ; \
apt-get install -y libblockdev-lvm-dev libblockdev-dev lvm2

RUN mkdir /iscsi-target-api
WORKDIR /iscsi-target-api

# COPY go.mod and go.sum files to the workspace
COPY go.mod .
COPY go.sum .


COPY . .
RUN if [ ! -d "/iscsi-target-api/vendor" ]; then  go mod vendor; fi
RUN make build-in-docker


FROM centos:7

RUN yum update -y ; \
yum install -y epel-release ; \
yum install -y scsi-target-utils libblockdev-lvm-devel.x86_64 libblockdev-devel.x86_64 \
yum clean all

RUN sed -i 's/tgtd_count=`pidof tgtd | wc -w`/tgtd_count\=1/g'  /usr/sbin/tgt-setup-lun

RUN sed -i 's/udev_rules = 1/udev_rules = 0/g' /etc/lvm/lvm.conf && \
    sed -i 's/udev_sync = 1/udev_sync = 0/g' /etc/lvm/lvm.conf

COPY --from=build /iscsi-target-api/bin/iscsi-target-api /
CMD ["/iscsi-target-api","-v","2", "logtostderr","true"]