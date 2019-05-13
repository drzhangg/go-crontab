package common

import (
	"encoding/json"
	"fmt"
)

//定时任务
type Job struct {
	Name     string `json:"name"`     //任务名
	Command  string `json:"command"`  //shell命令
	CronExpr string `json:"cronExpr"` //cron表达式
}

//http请求
type Response struct {
	ErrNo int         `json:"errNo"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//应答方法
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)

	response.ErrNo = errno
	response.Msg = msg
	response.Data = data

	//序列化json
	resp, err = json.Marshal(response)
	return
}

//反序列化job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)

	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		fmt.Println(err)
		return
	}

	ret = job
	return
}
