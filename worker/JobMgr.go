package worker

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type JobMgr struct {
	client *clientv3.Client `json:"client"`
	kv     clientv3.KV      `json:"kv"`
	lease  clientv3.Lease   `json:"lease"`
}

var (
	G_jobMgr *JobMgr
)

func InitJobMgr() (err error) {
	var (
		conf   clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	//初始化etcd配置
	conf = clientv3.Config{
		Endpoints:   G_config.EtcdEndPoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时
	}

	//新建etcd连接
	client, err = clientv3.New(conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	//获取etcd kv
	kv = clientv3.NewKV(client)

	//获取etcd租约
	lease = clientv3.NewLease(client)

	//赋值到单例
	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}

	return
}
