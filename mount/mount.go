package mount

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/term"
)

func main() {
    localDir := "/mnt/remote_folder"
    remoteDir := "//remote-ip/remote-folder"
    username := "myuser"
    password := "mypass"

    // 检查本地目录是否存在，如果不存在则创建
    _, err := os.Stat(localDir)
    if os.IsNotExist(err) {
        os.MkdirAll(localDir, 0755)
    }

    // 执行挂载命令
    cmd := exec.Command("mount", "-t", "cifs", remoteDir, localDir, "-o", fmt.Sprintf("username=%s,password=%s", username, password))
    err = cmd.Run()
	t,_:=term.MakeRaw(2)
	fmt.Printf("term:%v",t)
    if err != nil {
        fmt.Println("Failed to mount remote folder:", err)
        return
    }

    // 检查本地目录是否可用
    _, err = os.Stat(localDir)
    if err != nil {
        fmt.Println("Local folder is not accessible:", err)
        return
    }

    fmt.Println("Remote folder mounted successfully!")
}