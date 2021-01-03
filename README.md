# iscsi-target-api


## tgtd

```shell

$ tgt-admin --dump > /etc/tgt/targets.conf
$ tgt-admin -e 

# create volume 
$ tgtimg --op new --device-type disk --type disk --size 100m --file /var/lib/iscsi/test.img

$ tgt-setup-lun -n iqn.2017-07.com.hiroom2:debian-9  -d /var/lib/iscsi/test.img


$ tgtadm --lld iscsi --op new --mode target --tid 1 -T iqn.2017-07.com.hiroom2:debian-9
$ tgtadm --lld iscsi --op new --mode logicalunit --tid 1 --lun $i -b /var/lib/iscsi/10m-$i.img
```
