# iscsi-target-api



## Prerequisites

iSCSI manage tools are require with `iscsi-target-api`. Our goal is to support common iSCSI manage tools on Linux, currently tgtd supported is implemented.   

### tgtd

`iscsi-target-utils` must be installed properly. 

* If you use Ubuntu 16.04, install `tgt` package. Please visir [Ubuntu 16.04: Install tgt for iSCSI target
](https://www.hiroom2.com/2017/07/11/ubuntu-1604-tgt-en/) for more detail. 
    ```bash
    $ sudo apt install -y tgt
    ```
    
* If you use CentOS 7, install `scsi-target-utils` and configure SELinux. Please visit [CentOS 7: Install scsi-target-utils for iSCSI target
](https://www.hiroom2.com/2017/07/12/centos-7-scsi-target-utils-en/) for more deatil.

    ```bash
    $ sudo yum install -y epel-release
    $ sudo yum install -y scsi-target-utils
    $ sudo firewall-cmd --add-service=iscsi-target --permanent
    $ sudo firewall-cmd --reload
    $ sudo systemctl enable tgtd
    $ sudo systemctl restart tgtd
    ```
    
    ```bash
    $ cat <<EOF > tgtd-var-lib.te
    module tgtd-var-lib 1.0;
    
    require {
            type iscsi_var_lib_t;
            type tgtd_t;
            class file { open read write getattr };
            class dir { search };
    }
    
    #============= tgtd_t ==============
    allow tgtd_t iscsi_var_lib_t:file { open read write getattr };
    allow tgtd_t iscsi_var_lib_t:dir { search };
    EOF
    $ checkmodule -M -m -o tgtd-var-lib.mod tgtd-var-lib.te
    $ semodule_package -m tgtd-var-lib.mod -o tgtd-var-lib.pp
    $ sudo semodule -i tgtd-var-lib.pp
    $ rm -f tgtd-var-lib.te tgtd-var-lib.mod tgtd-var-lib.pp
    ```

## Build & Run

```bash
$ make build-img
$ make run-in-docker
```




## Curl Example

```bash
$ curl -XPOST -d '{"name":"test.img","size":"10m","path":"test"}' http://127.0.0.1:8811/createVol
$ curl -XPOST -d '{"targetIQN":"iqn.2017-07.k8s.default:myclaim", "volume": {"name":"test.img","path":"test"}}' http://127.0.0.1:8811/attachLun
$ curl -XDELETE -d '{"targetIQN":"iqn.2017-07.k8s.default:myclaim"}' http://127.0.0.1:8811/deleteTar
$ curl -XDELETE -d '{"name":"test.img","path":"test"}' http://127.0.0.1:8811/deleteVol
```


## Limitation

* One LUN per Target, volume is added at LUN `1`.

## TODO

* Support iSCSI ACL
* add API Authorization
* One target has multiple LUNs
    * One target represent on namespace, one LUN represent one PV . 
* Support `iscsitarget` manage tool
* Support LVM 
* Support volume thin provision 


## Reference
* tgtd CLI example

    ```shell
    
    $ tgt-admin --dump > /etc/tgt/targets.conf
    $ tgt-admin -e 
    
    # create volume 
    $ tgtimg --op new --device-type disk --type disk --size 100m --file /var/lib/iscsi/test.img
    
    $ tgt-setup-lun -n iqn.2017-07.com.hiroom2:debian-9  -d /var/lib/iscsi/test.img
    
    
    $ tgtadm --lld iscsi --op new --mode target --tid 1 -T iqn.2017-07.com.hiroom2:debian-9
    $ tgtadm --lld iscsi --op new --mode logicalunit --tid 1 --lun $i -b /var/lib/iscsi/10m-$i.img
    ```