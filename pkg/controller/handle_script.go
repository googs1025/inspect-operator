package controller

import (
	inspectv1alpha1 "github.com/myoperator/inspectoperator/pkg/apis/inspect/v1alpha1"
	"github.com/myoperator/inspectoperator/pkg/common"
	"github.com/myoperator/inspectoperator/pkg/workload/job"
	"github.com/myoperator/inspectoperator/pkg/workload/script"
	"k8s.io/klog/v2"
)

func handleScript(spec *inspectv1alpha1.InspectSpec) error {
	for _, task := range spec.Tasks {
		// 这里区分使用镜像还是用户自定义脚本
		if task.Task.Type == common.ScriptType {
			// 如果 Script 字段有值，默认使用用户定义的脚本内容
			if task.Task.Script != "" {
				// 固定的脚本镜像
				err := job.CreateJob(&task.Task, common.ScriptExecuteImage)
				if err != nil {
					klog.Error("create job err: ", err)
					return err
				}
			} else {
				// FIXME: 目前就是异步执行，后面controller就管不了
				go func() {
					err := script.RunLocalNode(task.Task.Source)
					if err != nil {
						klog.Error("create script err: ", err)
						return
					}
				}()
			}
		}
	}
	return nil
}
