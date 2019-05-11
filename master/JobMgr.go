package master

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/drzhangg/go-crontab/common"
	"time"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client `json:"client"`
	kv     clientv3.KV      `json:"kv"`
	lease  clientv3.Lease   `json:"lease"`
}

var (
	G_jobMgr *JobMgr
)

//初始化管理器
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
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时时间
	}

	//连接etcd
	client, err = clientv3.New(conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	//获取kv
	kv = clientv3.NewKV(client)

	//获取租约
	lease = clientv3.NewLease(client)

	//赋值单例
	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}


//保存接口
func (jobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobKey   string
		jobValue []byte
		putResp  *clientv3.PutResponse
		oldJobObj common.Job
	)

	//etcd的保存key
	jobKey = "/cron/jobs/" + job.Name

	//json解析job
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}

	//通过kv保存jobValue到etcd，并且查询保存前的值
	putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return
	}

	//如果保存前有值，返回旧值
	if putResp.PrevKv != nil {
		if err = json.Unmarshal([]byte(putResp.PrevKv.Value), &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}
