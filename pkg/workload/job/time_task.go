package job

import (
	"context"
	"fmt"
	"github.com/myoperator/inspectoperator/pkg/request"
	. "github.com/myoperator/inspectoperator/pkg/workload"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"time"
)

// Fn 定义函数类型
type Fn func(taskName string, t string) (bool, error, string)

// MyTickerTask 定时器
type MyTickerTask struct {
	MyTick   *time.Ticker
	Runner   Fn
	TaskName string
	TaskType string
	stopC    chan struct{}
}

func NewTickerTask(interval int, f Fn, taskName string, taskType string) *MyTickerTask {
	return &MyTickerTask{
		MyTick:   time.NewTicker(time.Duration(interval) * time.Second),
		Runner:   f,
		TaskName: taskName,
	}
}

// Start 启动定时器需要执行的任务
func (t *MyTickerTask) Start() {
	for {
		select {
		case <-t.MyTick.C:
			// 间隔时间就调用 t.Runner，并判断是否退出循环
			isReStart, err, res := t.Runner(t.TaskName, t.TaskType)
			if isReStart {
				continue
			}

			if err != nil || isReStart != true {
				klog.Info("发送job完成的消息")
				// FIXME: 这里不能写死。
				request.Post("http://42.193.17.123:31130/v1/send", t.TaskName, fmt.Sprintf("you job res: %v", res))
				t.stopC <- struct{}{}
			}
		case <-t.stopC:
			klog.Info("stop the cron task...")
			time.Sleep(time.Second * 3)
			return
		}
	}
}

// GetJobStatus 获取job状态
// 返回值：bool:代表是否还要继续执行，error:是否有错误，string:代表结果
func GetJobStatus(taskName string, t string) (bool, error, string) {
	res := getJobTaskName(taskName, t)

	getJob, err := ClientSet.BatchV1().Jobs("default").Get(context.Background(), res, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		klog.Error("not found error ", err)
		return false, err, fmt.Sprintf("not found error")
	} else if err != nil {
		klog.Error("get job error: ", err)
		return false, err, err.Error()
	}
	klog.Info("get job: ", getJob.Name)
	res, isReStart := checkStatus(&getJob.Status)
	return isReStart, nil, res
}

// 检查 status 返回值：string 代表结果，bool 代表是否还要继续定时执行
func checkStatus(status *batchv1.JobStatus) (string, bool) {
	if status.Succeeded == 1 {
		return "succeeded", false
	} else if status.Failed == 1 {
		return "failed", false
	}
	return "", true

}
