package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	cmdStr := strings.Join(os.Args, " ")
	fmt.Printf("命令:%s \n", cmdStr)
	cmd := exec.Command(os.Args[1], os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
}
