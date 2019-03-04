package main

import (
	"fmt"
	"os/exec"
)

func main() {
	var (
		cmd *exec.Cmd
		output []byte
		err error
	)

	cmd = exec.Command("/bin/bash","-c","echo helloworld")

	if output,err = cmd.CombinedOutput();err != nil{
		fmt.Println(err)
		return
	}

	fmt.Println(string(output))
}
