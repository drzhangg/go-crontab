package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

/**
	lease租约实现kv过期
 */
func main() {
	var(
		config clientv3.Config
		client *clientv3.Client
		err error
		lease clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
		kv clientv3.KV
		putResp *clientv3.PutResponse
		getResp *clientv3.GetResponse
		keepResp *clientv3.LeaseKeepAliveResponse
		KeepRespChan <- chan *clientv3.LeaseKeepAliveResponse
	)

	//配置文件
	config = clientv3.Config{
		Endpoints:[]string{"47.99.240.52:2379"},//集群列表
		DialTimeout:5 * time.Second,	//设置超时时间
	}

	//创建一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//申请一个租约
	lease = clientv3.NewLease(client)

	//申请一个10秒的租约
	if leaseGrantResp,err = lease.Grant(context.TODO(),10);err != nil {
		fmt.Println(err)
		return
	}

	//拿到租约的ID
	leaseId = leaseGrantResp.ID

	//自动续租
	if KeepRespChan,err = lease.KeepAlive(context.TODO(),leaseId);err != nil {
		fmt.Println(err)
		return
	}

	//处理续租应答的协程
	go func() {
		for{
			select {
			case keepResp =  <- KeepRespChan:
				if KeepRespChan == nil {
					fmt.Println("租约已到期")
					goto END
				}else {	//	每秒会续租一次，所有就会收到一次应答
					fmt.Println("收到自动续租应答：",keepResp.ID)
				}

			}
			END:
		}
	}()

	//获得kv API子集
	kv = clientv3.NewKV(client)

	//put一个kv，让它与租约关联起来，从而实现10秒后	自动过期
	if putResp,err = kv.Put(context.TODO(),"/cron/lock/job1","",clientv3.WithLease(leaseId));err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("写入成功：",putResp.Header.Revision)

	//定时查看一下key过期了没有
	for{
		if getResp,err = kv.Get(context.TODO(),"/cron/lock/job1");err != nil {
			fmt.Println(err)
			return
		}

		if getResp.Count == 0 {
			fmt.Println("过期了")
			break
		}
		fmt.Println("还没过期：",getResp.Kvs)
		time.Sleep(2 * time.Second)
	}


}
