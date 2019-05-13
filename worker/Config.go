package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//配置列表
type Config struct {
	EtcdEndPoints   []string `json:"etcdEndPoints"`		//etcd集群列表
	EtcdDialTimeout int `json:"etcdDialTimeout"`	//etcd超时时间
}

var (
	G_config *Config
)

func InitConfig(filename string) (err error) {
	var(
		content []byte
		conf Config
	)

	//1.读取配置文件
	content,err = ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("readFile failed:",err)
		return
	}

	//2.反序列化读取的配置文件
	err = json.Unmarshal(content,&conf)
	if err != nil {
		return
	}

	//3.赋值单例
	G_config = &conf

	return
}
