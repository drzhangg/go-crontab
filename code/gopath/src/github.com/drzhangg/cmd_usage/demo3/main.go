package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type result struct {
	err error
	output []byte
}

func main() {

	//执行一个cmd，让它在一个协程里执行，让它执行2秒：sleep 2;echo hello;

	var(
		ctx context.Context
		cancelFunc context.CancelFunc
		cmd *exec.Cmd
		resultChan chan *result
		res *result
	)

	//创建一个结果队列
	resultChan = make(chan *result,1000)

	ctx,cancelFunc = context.WithCancel(context.TODO())

	go func() {
		var (
			output []byte
			err error
		)

		cmd = exec.CommandContext(ctx,"/bin/bash","-c","sleep 2;echo hello;")

		//执行任务，捕获输出
		output,err = cmd.CombinedOutput()

		//把任务输出结果，传给main协程
		resultChan <- &result{
			err:err,
			output:output,
		}
	}()

	time.Sleep(1 * time.Second)

	//取消上下文
	cancelFunc()

	res = <- resultChan

	fmt.Println(res.output,res.err)
}
