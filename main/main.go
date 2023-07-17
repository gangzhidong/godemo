package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	cmdStr := strings.Join(os.Args, " ")
	fmt.Printf("全部命令:%s \n", cmdStr)
	fmt.Printf("命令:%s \n", os.Args[1])
	fmt.Printf("参数:%s \n", strings.Join(os.Args[1:], " "))
	cmd := exec.Command(os.Args[1], os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
}
