package master

import (
	"encoding/json"
	"fmt"
	"go-corntab/common"
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

//初始化http配置
func InitApiServer() (err error) {
	var (
		mux      *http.ServeMux
		listener net.Listener

		httpServer    *http.Server
		staticDir     http.Dir     //静态文件根目录
		staticHandler http.Handler //静态文件的HTTP回调
	)

	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)     //保存任务
	mux.HandleFunc("/job/delete", handleJobDelete) //删除任务
	mux.HandleFunc("/job/list", handleJobList)     //获取全部job
	mux.HandleFunc("/job/kill", handleJobKill)     //杀死指定job
	mux.HandleFunc("/job/log", handleJobLog)       //日志监听保存接口
	mux.HandleFunc("/worker/list", handleWorkerList)

	//配置静态文件目录
	staticDir = http.Dir(G_config.WebRoot)
	staticHandler = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

	//启动tcp监听
	listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort))
	if err != nil {
		fmt.Println("listener:", err)
		return
	}

	//创建一个HTTP服务
	httpServer = &http.Server{
		ReadTimeout:  time.Millisecond * time.Duration(G_config.ApiReadTimeout),
		WriteTimeout: time.Millisecond * time.Duration(G_config.ApiWriteTimeout),
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

	//1.解析post表单
	err = req.ParseForm()
	if err != nil {
		fmt.Println("parseForm:", err)
		goto ERR
	}

	//2.取表单中的job字段
	postJob = req.PostForm.Get("job")

	//3.反序列化job
	err = json.Unmarshal([]byte(postJob), &job)
	if err != nil {
		fmt.Println("unmarshal:", err)
		goto ERR
	}

	//4.保存数据到etcd
	old, err = G_jobMgr.SaveJob(&job)
	if err != nil {
		fmt.Println("errr:", err)
		goto ERR
	}
	fmt.Println("SaveJob:", old)

	//5.请求成功，返回消息
	re, err = common.BuildResponse(0, "success", old)
	if err == nil {
		fmt.Println("rebuildResp:", err)
		resp.Write(re)
	}
	fmt.Println("rebuildResp1:", err)
	return
ERR:
	//请求失败，返回错误信息
	re, err = common.BuildResponse(-1, err.Error(), nil)
	if err == nil {
		fmt.Println("build failed:", err)
		resp.Write(re)
	}
}

//删除任务
//POST /job/delete name=job1
func handleJobDelete(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		name   string
		oldJob *common.Job
		resp   []byte
	)

	//POST:
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	//删除的任务名
	name = r.PostForm.Get("name")

	//通过etcd删除指定任务
	oldJob, err = G_jobMgr.DeleteJob(name)
	if err != nil {
		goto ERR
	}

	resp, err = common.BuildResponse(0, "success", oldJob)
	if err == nil {
		w.Write(resp)
	}
	return
ERR:
	resp, err = common.BuildResponse(-1, err.Error(), nil)
	if err == nil {
		w.Write(resp)
	}
}

//获取jobList
//GET /job/list
func handleJobList(w http.ResponseWriter, r *http.Request) {

	var (
		err     error
		jobList []*common.Job
		resp    []byte
	)

	//读取全部任务
	jobList, err = G_jobMgr.ListJobs()
	if err != nil {
		fmt.Println("get JobList failed:", err)
		goto ERR
	}

	//请求成功，返回请求
	resp, err = common.BuildResponse(0, "success", jobList)
	if err == nil {
		w.Write(resp)
	}

	return
ERR:
	//请求失败
	resp, err = common.BuildResponse(-1, err.Error(), nil)
	if err == nil {
		w.Write(resp)
	}
}

//杀死任务
//POST /job/kill name=job1
func handleJobKill(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		jobName string
		bytes   []byte
	)

	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	jobName = r.PostForm.Get("name")

	//杀死一个任务
	err = G_jobMgr.KillJob(jobName)
	if err != nil {
		goto ERR
	}

	bytes, err = common.BuildResponse(0, "success", nil)
	if err == nil {
		w.Write(bytes)
	}

	return
ERR:
	bytes, err = common.BuildResponse(-1, err.Error(), nil)
	if err == nil {
		w.Write(bytes)
	}
}

//查询任务日志
func handleJobLog(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		name       string //任务名字
		skipParam  string //从第几条开始
		limitParam string //返回多少条
		skip       int
		limit      int
		logArr     []*common.JobLog
		bytes      []byte
	)

	//解析GET参数
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	//获取请求参数 /job/log?name=job10&skip=0&limit=10
	name = req.Form.Get("name")
	skipParam = req.Form.Get("skip")
	limitParam = req.Form.Get("limit")
	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}

	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}

	if logArr, err = G_logMgr.ListLog(name, skip, limit); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", logArr); err == nil {
		resp.Write(bytes)
	}

	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

//获取健康worker节点列表
func handleWorkerList(w http.ResponseWriter, r *http.Request) {
	var (
		workerArr []string
		err       error
		bytes     []byte
	)
	if workerArr, err = G_workerMgr.ListWorkers(); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", workerArr); err == nil {
		w.Write(bytes)
	}

	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}
