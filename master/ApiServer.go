package master

import (
	"encoding/json"
	"fmt"
	"github.com/drzhangg/go-crontab/common"
	"net"
	"net/http"
	"strconv"
	"time"
)

//任务的HTTP接口
type ApiServer struct {
	httpServer *http.Server
}

var (
	G_apiServer *ApiServer
)

//保存任务接口
//POST job={"name":"job1","command":"echo hello","cronExpr":"* * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		old     *common.Job
		re      []byte
	)

	fmt.Println("resp:",resp)
	fmt.Println("req:",req)
	//1.解析post表单
	err = req.ParseForm()
	if err != nil {
		fmt.Println(err)
		goto ERR
	}

	//2.取表单中的job字段
	postJob = req.PostForm.Get("job")

	//3.反序列化job
	err = json.Unmarshal([]byte(postJob), &job)
	if err != nil {
		fmt.Println(err)
		goto ERR
	}

	//4.保存数据到etcd
	old, err = G_jobMgr.SaveJob(&job)
	if err != nil {
		goto ERR
	}

	//5.请求成功，返回消息
	re, err = common.BuildResponse(0, "success", old)
	if err != nil {
		resp.Write(re)
	}
	return
ERR:
	//请求失败，返回错误信息
	re, err = common.BuildResponse(-1, err.Error(), nil)
	if err != nil {
		resp.Write(re)
	}
}

//初始化http配置
func InitApiServer() (err error) {
	var (
		mux      *http.ServeMux
		listener net.Listener

		httpServer *http.Server
	)

	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/hello", HandleHello)
	mux.HandleFunc("/job/save", handleJobSave)

	//启动tcp监听
	listener, err = net.Listen("tcp", ":" + strconv.Itoa(G_config.ApiPort))
	if err != nil {
		fmt.Println(err)
		return
	}

	//创建一个HTTP服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout),
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout),
		Handler:      mux,
	}

	//赋值单例
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	//启动服务端
	go httpServer.Serve(listener)

	return
}

func HandleHello(w http.ResponseWriter, r *http.Request) {

	fmt.Println("resp:",w)
	fmt.Println("req:",r)
	w.Write([]byte("hello test"))
}
