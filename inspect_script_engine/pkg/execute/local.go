package execute

import (
	"k8s.io/klog/v2"
	"os/exec"
)

// RunLocalNode 执行本地节点命令
// FIXME: 不是在宿主机上执行，而是在容器中执行
func (sc *ScriptExecutor) RunLocalNode() error {
	// 修正镜像没有bash
	cmd := exec.Command("sh", sc.Path)

	cmd.Stdout = &sc.StdOut // 标准输出
	cmd.Stderr = &sc.StdErr // 标准错误
	err := cmd.Run()
	if err != nil {
		klog.Error("cmd.Run() failed with: ", err)
		return err
	}
	return nil
}
