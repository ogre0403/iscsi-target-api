# iscsi-target-api


## tgtd Cli example

```shell

$ tgt-admin --dump > /etc/tgt/targets.conf
$ tgt-admin -e 

# create volume 
$ tgtimg --op new --device-type disk --type disk --size 100m --file /var/lib/iscsi/test.img

$ tgt-setup-lun -n iqn.2017-07.com.hiroom2:debian-9  -d /var/lib/iscsi/test.img


$ tgtadm --lld iscsi --op new --mode target --tid 1 -T iqn.2017-07.com.hiroom2:debian-9
$ tgtadm --lld iscsi --op new --mode logicalunit --tid 1 --lun $i -b /var/lib/iscsi/10m-$i.img
```
## Curl Example

```bash
$ curl -XPOST -d '{"name":"aaa.img","size":"111m"}' http://140.110.30.57/createVol
$ curl -XPOST -d '{"targetIQN":"iqn.2017-07.com.hiroom2:aaadd", "volume": {"name":"aaa.img"}}' http://140.110.30.57/attachLun
$ curl -XDELETE -d '{"targetIQN":"iqn.2017-07.com.hiroom2:aaadd"}' http://140.110.30.57/deleteTar
$ curl -XDELETE -d '{"name":"aaa.img"}' http://140.110.30.57/deleteVol
```


## PV / PVC mapping

```bash
iqn.2017-07.k8s.<namespace>:<PVC-UUID>
```

## Limitation

* One LUN per Target, volume is added at LUN `1`.

## TODO

* Support iSCSI ACL
* add API Authorization
* One target has multiple LUNs
    * One target represent on namespace, one LUN represent one PV . 
* Support `iscsitarget`


