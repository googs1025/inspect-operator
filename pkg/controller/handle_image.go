package controller

import (
	inspectv1alpha1 "github.com/myoperator/inspectoperator/pkg/apis/inspect/v1alpha1"
	"github.com/myoperator/inspectoperator/pkg/common"
	"github.com/myoperator/inspectoperator/pkg/workload/job"
	"k8s.io/klog/v2"
)

// handleImage 处理镜像巡检任务
// 创建 Job 执行任务
func handleImage(spec *inspectv1alpha1.InspectSpec) error {

	for _, task := range spec.Tasks {
		if task.Task.Type == common.ImageType {
			err := job.CreateJob(&task.Task, task.Task.Source)
			if err != nil {
				klog.Error("create job err: ", err)
				return err
			}
			// 启动定时任务
			ticker := job.NewTickerTask(30, job.GetJobStatus, task.Task.TaskName, task.Task.Type)
			klog.Info("start cron task...")
			go ticker.Start()
		}
	}
	return nil
}
