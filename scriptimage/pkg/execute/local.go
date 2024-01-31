package execute

import (
	"bytes"
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"scriptimage/pkg/common"
	"scriptimage/pkg/request"
)

func RunLocalNode() error {
	path := common.GetWd()
	// 修正镜像没有bash
	cmd := exec.Command("sh", path+common.ScriptFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // 标准输出
	cmd.Stderr = &stderr // 标准错误
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	klog.Info("\nout:\n", outStr, "err:\n", errStr)
	if err != nil {
		klog.Error("cmd.Run() failed with: ", err)
		request.Post(os.Getenv("message-operator-url"),
			fmt.Sprintf("execute bash script"), fmt.Sprintf("res: %v, err: %v", outStr, errStr))
		return err
	}
	klog.Info("finish to send the script message...")
	// FIXME: 这里不能写死。
	//request.Post(os.Getenv("message-operator-url"),
	//	fmt.Sprintf("execute bash script"), fmt.Sprintf("res: %v, err: %v", outStr, errStr))

	return nil
}
