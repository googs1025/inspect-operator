package main

import (
	"k8s.io/klog/v2"
	"log"
	"os"
	"path"
	"scriptimage/pkg/common"
	"scriptimage/pkg/execute"
	"scriptimage/pkg/manager"
	"scriptimage/pkg/manager/hook/log_hook"
)

/*
	使用更流程的方式实现，参考hook 实现
*/

func main() {

	sc := execute.NewScriptExecutor(
		path.Join(common.GetWd(), common.ScriptFile),
		os.Getenv("script"),
		os.Getenv("taskName"),
		os.Getenv("script_location"),
		execute.NewInfo(os.Getenv("user"), os.Getenv("password"), os.Getenv("ip")),
	)

	workflowManager := manager.NewWorkflowManager(sc)

	// 创建一个日志记录钩子对象
	logger := log.New(log.Writer(), "", log.LstdFlags)
	loggingHook := &log_hook.LoggingHook{
		Logger: logger,
	}

	// 注册日志记录钩子
	workflowManager.RegisterHook(loggingHook)

	// TODO: 注册通知钩子

	// 执行
	err := workflowManager.Start()
	if err != nil {
		klog.Error("script executor error: ", err)
		return
	}

	klog.Info("script executor successful")
}
