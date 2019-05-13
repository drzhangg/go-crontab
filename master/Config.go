package master

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//程序配置
type Config struct {
	ApiPort         int      `json:"apiPort"`
	ApiReadTimeout  int      `json:"apiReadTimeout"`
	ApiWriteTimeout int      `json:"apiWriteTimeout"`
	EtcdEndPoints   []string `json:"etcdEndPoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	WebRoot         string   `json:"webroot"`
}

var (
	G_config *Config
)

//加载配置
func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Config
	)

	//1.读取配置文件
	content, err = ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	//2.反序列化读取的文件
	err = json.Unmarshal(content, &conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	//3.赋值单例
	G_config = &conf

	fmt.Println(G_config)

	return
}
