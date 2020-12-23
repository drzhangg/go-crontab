package worker

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go-corntab/common"
	"time"
)

type JobMgr struct {
	client  *clientv3.Client `json:"client"`
	kv      clientv3.KV      `json:"kv"`
	lease   clientv3.Lease   `json:"lease"`
	watcher clientv3.Watcher `json:"watcher"`
}

var (
	G_jobMgr *JobMgr
)

func InitJobMgr() (err error) {
	var (
		conf    clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
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

	//监听
	watcher = clientv3.NewWatcher(client)

	//赋值到单例
	G_jobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}

	//启动任务监听
	G_jobMgr.watchJobs()

	//启动监听killer
	G_jobMgr.watchKiller()

	return
}

//监听任务变化
func (jobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvpair             *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobEvent           *common.JobEvent
		jobName            string
	)

	//1.get一下/cron/jobs/目录下的所有任务，并且获知当前集群的revision
	getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix())
	if err != nil {
		fmt.Println(err)
		return
	}

	//获取当前有哪些任务
	for _, kvpair = range getResp.Kvs {
		//反序列化json得到job
		if job, err = common.UnpackJob(kvpair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)

			//把这个job同步给scheduler（调度协程）
			G_scheduler.PushJobEvent(jobEvent)
			fmt.Println(*jobEvent)
		}
	}

	//2.从该revision向后监听变化事件
	go func() {
		//从get的后续版本开始监听变化
		watchStartRevision = getResp.Header.Revision + 1

		//监听/cron/jobs/目录的后续变化
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		//处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //任务保存事件
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					//构建一个更新Event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)

				case mvccpb.DELETE: //任务删除事件
					jobName = common.ExtractJobName(string(watchEvent.Kv.Value))

					job = &common.Job{Name: jobName}
					//构建一个删除Event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}

				//推送给scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}

	}()

	return
}

//创建任务执行锁
func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
	return
}
