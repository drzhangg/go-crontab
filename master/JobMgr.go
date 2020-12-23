package master

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go-corntab/common"
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
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj common.Job
	)

	//etcd的保存key
	jobKey = common.JOB_SAVE_DIR + job.Name

	fmt.Println("key:", jobKey)

	//json解析job
	if jobValue, err = json.Marshal(job); err != nil {
		fmt.Println("Marshal:", err)
		return
	}

	fmt.Println("jobValue:", string(jobValue))

	//通过kv保存jobValue到etcd，并且查询保存前的值
	putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println("put success:", putResp)

	//如果保存前有值，返回旧值
	if putResp.PrevKv != nil {
		if err = json.Unmarshal([]byte(putResp.PrevKv.Value), &oldJobObj); err != nil {
			fmt.Println("err:", err)
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

//删除job
func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		jobName   string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)

	//获取etcd中保存的key
	jobName = common.JOB_SAVE_DIR + name

	//通过获取的key删除etcd
	delResp, err = jobMgr.kv.Delete(context.TODO(), jobName, clientv3.WithPrevKV())
	if err != nil {
		fmt.Println(err)
		return
	}

	//返回被删除的信息
	if len(delResp.PrevKvs) != 0 {
		//解析被删除的旧值，并返回
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}

		oldJob = &oldJobObj
	}
	return
}

//列举全部任务
func (jobMgr *JobMgr) ListJobs() (jobLists []*common.Job, err error) {
	var (
		jobKey  string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *common.Job
	)

	//获取etcd的目录key
	jobKey = common.JOB_SAVE_DIR

	//获取父目录下的所有任务
	getResp, err = jobMgr.kv.Get(context.TODO(), jobKey, clientv3.WithPrefix())
	if err != nil {
		fmt.Println("getResp list failed:", err)
		return
	}

	jobLists = make([]*common.Job, 0)

	//遍历所有任务，进行反序列化
	for _, kvPair = range getResp.Kvs {
		job = &common.Job{}
		//json.Unamrshal(k,v)   k是读取到的json字符串，要解析的；v是要讲json解析成的类型
		if err = json.Unmarshal(kvPair.Value, &job); err != nil {
			err = nil
			continue
		}

		jobLists = append(jobLists, job)
	}
	return
}

//杀死任务
func (jobMgr *JobMgr) KillJob(name string) (err error) {
	var (
		killKey        string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)

	//获取job的key
	killKey = common.JOB_KILLER_DIR + name

	//让worker监听一个put操作，创建一个租约让其稍后自动过期即可
	//创建一个租约
	leaseGrantResp, err = jobMgr.lease.Grant(context.TODO(), 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	//获取租约id
	leaseId = leaseGrantResp.ID

	/**
	这里其实是每次put进etcd中一个空值，然后给他一个1秒的租约，当租约过期时，这个etcd就会被delete，
	以此达到kill的效果
	*/
	//设置kill标记
	_, err = jobMgr.kv.Put(context.TODO(), killKey, "", clientv3.WithLease(leaseId))
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
