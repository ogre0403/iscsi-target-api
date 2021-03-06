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

* Create Volume
    * tgtimg volume
        ```bash
        $ curl -XPOST \
          -d '{"type":"tgtimg","name":"test.img","size":10,"unit":"MiB","group":"test"}' \
          --user admin:password \
          http://127.0.0.1:8811/createVol
        ```
   
    * LVM volume
        ```bash
        $ curl -XPOST \
        -d '{"type":"lvm","name":"test","size":8,"unit":"MiB","group":"vg-0"}' \
        --user admin:password \
        http://127.0.0.1:8811/createVol
        ```
   
* Create Thin provision Volume

    * LVM Thin provision volume
        
        **Note:** LVM thin provision needs pool parameter, this is defined in `iscsi-target-api` parameter.
        
        ```bash
        $ curl -XPOST \
        -d '{"type":"lvm","name":"test","size":8,"unit":"MiB","group":"vg-0", "thin":true}' \
        --user admin:password \
        http://127.0.0.1:8811/createVol
        ```   
    * tgtimg thin provision volume
        ```bash
        $ curl -XPOST \
          -d '{"type":"tgtimg","name":"test.img","size":10,"unit":"MiB","group":"test","thin":true}' \
          --user admin:password \
          http://127.0.0.1:8811/createVol
        ```   
* Create target & LUN      
    ```
    $ curl -XPOST \
      -d '{"targetIQN":"iqn.2017-07.k8s.default:myclaim", "volume": {"type":"tgtimg","name":"test.img","group":"test"}}' \
      --user admin:password \
      http://127.0.0.1:8811/attachLun
    ```
* Delete target
    ```
    $ curl -XDELETE \
      -d '{"targetIQN":"iqn.2017-07.k8s.default:myclaim"}' \
      --user admin:password \
      http://127.0.0.1:8811/deleteTar
    ```

* Delete Volume

    ```
    $ curl -XDELETE \
      -d '{"type":"tgtimg","name":"test.img","group":"test"}' \
      --user admin:password \
      http://127.0.0.1:8811/deleteVol
    ```

## Json body

* Volume
    ```json
    {
      "type": "tgtimg", 
      "group": "test",
      "name": "test.img",
      "size": 10,
      "unit": "MiB",
      "thin": false
    }
    ```

* Target 
    ```json
    {
      "targetIQN": "iqn.2017-07.k8s.default:myclaim", 
      "volume": {
          "type": "tgtimg",
          "group": "test",
          "name": "test.img"
      },
      "aclList": ["192.168.1.0/24"],
      "enableCHAP": false
    }
    ```

## Limitation

* One LUN per Target, volume is added at LUN `1`.

## TODO
* One target has multiple LUNs
    * One target represent on namespace, one LUN represent one PV . 
* Support `iscsitarget` manage tool
* ~~Support iSCSI ACL~~
    * ~~CHAP~~
    * ~~Initiator IP~~
* ~~add API Authorization~~
* ~~Support LVM~~ 
* ~~Support volume thin provision~~ 


## Reference
* tgtd CLI example

    ```shell
    
    $ tgt-admin --dump > /etc/tgt/targets.conf
    $ tgt-admin -e 
    
    # create volume 
    $ tgtimg --op new --device-type disk --type disk --size 100m --file /var/lib/iscsi/test.img
    
    $ tgt-setup-lun -n iqn.2017-07.com.hiroom2:debian-9  -d /var/lib/iscsi/test.img
    
    # thin provision
    $ tgtadm --lld iscsi --mode logicalunit --op update --tid 1 --lun 1 --params thin_provisioning=1
    
    # find tid first
    $ tgtadm --lld iscsi --op new --mode target --tid 1 -T iqn.2017-07.com.hiroom2:debian-9
    $ tgtadm --lld iscsi --op new --mode logicalunit --tid 1 --lun 1 -b /var/lib/iscsi/10m-$i.img
    
    # setup ACL
    # accept all 
    $ tgtadm --lld iscsi --op bind --mode target --tid 1 -I ALL
    # accept ip
    $ tgtadm --lld iscsi --op bind --mode target --tid 1 -I 192.168.1.1
  
    # remove ACL 
    $ tgtadm --lld iscsi --op unbind --mode target --tid 1 -I 192.168.1.1
  
    # setup CHAP
    $ tgtadm --lld iscsi --op bind --mode account --tid 1 --user benjr --outgoing
    $ tgtadm --lld iscsi --op bind --mode account --tid 1 --user benjr
    
    # remove CHAP
    $ tgtadm --lld iscsi --op unbind --mode account --tid 1 --user benjr --outgoing
    $ tgtadm --lld iscsi --op unbind --mode account --tid 1 --user benjr
    
    ```