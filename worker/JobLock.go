package worker

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/drzhangg/go-crontab/common"
)

//分布式锁（通过抢占一个TXN事务，谁先抢到谁就占到了🔐）
type JobLock struct {
	kv    clientv3.KV
	lease clientv3.Lease

	jobName    string
	cancelFunc context.CancelFunc //用于终止自动续租
	leaseId    clientv3.LeaseID   //租约ID
	isLocked   bool               //是否上锁成功
}

//初始化一把🔐
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

//尝试上锁
func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)
	//1.创建租约（5秒）
	leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5)
	if err != nil {
		fmt.Println(err)
		return
	}

	//context用于取消自动续租
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	//获取租约id
	leaseId = leaseGrantResp.ID

	//2.自动续租
	keepRespChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseId)
	if err != nil {
		goto FAIL
	}

	//3.处理续租应答的协程
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)

		for {
			select {
			case keepResp = <-keepRespChan: //自动续租的应答
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	//4.创建事务txn
	txn = jobLock.kv.Txn(context.TODO())

	//锁路径
	lockKey = common.JOB_LOCK_DIR + jobLock.jobName

	//5.事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	//提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	//6.成功返回，失败释放租约
	if !txnResp.Succeeded { //锁被占用
		err = common.ERR_LOCK_ALERADY_REQUIRED
		goto FAIL
	}

	//抢锁成功
	jobLock.leaseId = leaseId
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true
	return

FAIL:
	cancelFunc()                                  //取消自动续租
	jobLock.lease.Revoke(context.TODO(), leaseId) //释放租约
	return
}

//释放锁
func (jobLock *JobLock) Unlock() {
	if jobLock.isLocked {
		jobLock.cancelFunc()                                  //取消程序自动续租的协程
		jobLock.lease.Revoke(context.TODO(), jobLock.leaseId) //释放租约
	}
}
