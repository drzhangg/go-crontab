package worker

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/drzhangg/go-crontab/common"
	"net"
	"time"
)

//注册节点到etcd：/cron/workers/IP地址
type Register struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	localIP string //本机IP
}

var (
	G_register *Register
)

func InitRegister() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIP string
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndPoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	//建立etcd连接
	client, err = clientv3.New(config)
	if err != nil {
		return
	}

	//申请一个kv
	kv = clientv3.NewKV(client)
	//申请一个租约
	lease = clientv3.NewLease(client)

	//本机IP
	if localIP, err = getLocalIP(); err != nil {
		return
	}

	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIP,
	}

	//服务注册
	go G_register.keepOnline()

	return
}

//获取本机网卡IP
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet //ip地址
		isIpNet bool
	)

	//获取所有网卡
	addrs, err = net.InterfaceAddrs()
	if err != nil {
		return
	}

	//取第一个非io的网卡
	for _, addr = range addrs {
		//这个网络地址是ip地址:ipv4或ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet {
			//跳过ipv6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}

	err = common.ERR_NO_LOCAL_IP_FOUND
	return
}

//注册到/cron/workers/IP，并自动续租
func (register *Register) keepOnline() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)

	for {
		//注册路径
		regKey = common.JOB_WORKER_DIR + register.localIP

		cancelFunc = nil

		//创建10秒的租约
		leaseGrantResp, err = register.lease.Grant(context.TODO(), 10)
		if err != nil {
			goto RETRY
		}

		//自动续租
		keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID)
		if err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		//注册到etcd
		_, err = register.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID))
		if err != nil {
			goto RETRY
		}

		//处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { //续租失败
					goto RETRY
				}
			}
		}
	RETRY:
		time.Sleep(time.Second * 1)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}
