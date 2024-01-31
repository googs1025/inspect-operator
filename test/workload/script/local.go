package script

import (
	"bytes"
	"fmt"
	"github.com/myoperator/inspectoperator/pkg/common"
	"github.com/myoperator/inspectoperator/test/request"
	"k8s.io/klog/v2"
	"os/exec"
)

func RunLocalNode(script string) error {
	path := common.GetWd()
	// 修正镜像没有bash
	cmd := exec.Command("sudo sh", path+"/script/"+script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // 标准输出
	cmd.Stderr = &stderr // 标准错误
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	klog.Info("out: %s\nerr: %s\n", outStr, errStr)
	if err != nil {
		klog.Error("cmd.Run() failed with %s\n", err)
		return err
	}
	klog.Info("finish to send the script message...")
	// FIXME: 这里不能写死。
	request.Post("http://42.193.17.123:31130/v1/send",
		fmt.Sprintf("bash script name: %v", script), fmt.Sprintf("res: %v, err: %v", outStr, errStr))

	return nil
}
