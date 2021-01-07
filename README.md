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

## PV / PVC mapping

```bash
iqn.2017-07.k8s.<namespace>:<PVC-UUID>
```

## Limitation

* One LUN per Target, vloume is add at LUN `1`.

## TODO

* One target has multiple LUNs
    * One target represent on namespace, one LUN represent one PV . 

* Support `iscsitarget`


## Curl Example

```bash
$ curl -XPOST -d '{"name":"aaa.img","size":"111m"}' http://140.110.30.57/createVol
$ curl -XPOST -d '{"targetIQN":"iqn.2017-07.com.hiroom2:aaadd", "volume": {"name":"aaa.img"}}' http://140.110.30.57/attachLun
```
