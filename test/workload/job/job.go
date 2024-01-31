package job

import (
	"context"
	"encoding/json"
	"fmt"
	inspectv1alpha1 "github.com/myoperator/inspectoperator/pkg/apis/inspect/v1alpha1"
	"github.com/myoperator/inspectoperator/pkg/common"
	. "github.com/myoperator/inspectoperator/test/workload"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

// CreateJob
// @see https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/job-v1/#create-create-a-job
func CreateJob(task *inspectv1alpha1.Task, image string) error {

	job := jobSpec(task, image)

	j, _ := json.Marshal(job)
	klog.Info("Create jobs with params: ", string(j))

	// create cronjob
	jobResult, err := ClientSet.BatchV1().Jobs("default").Create(context.TODO(), job, metav1.CreateOptions{})
	// FIXME: 处理已经存在job的错误 errors包
	// 方案：捞出exists错误，先删除原job，再创建。
	if err != nil && errors.IsAlreadyExists(err) {
		// 先删除原本的job
		klog.Error("Error maybe is AlreadyExists error: ", err)
		foreground := metav1.DeletePropagationForeground
		deleteOptions := metav1.DeleteOptions{PropagationPolicy: &foreground}
		if err = ClientSet.BatchV1().Jobs("default").Delete(context.Background(), getJobTaskName(task.TaskName, task.Type), deleteOptions); err != nil {
			klog.Error("Delete job error: ", err)
			return err
		}
		// NOTE: 需要等待删除才执行create
		time.Sleep(time.Second * 60)
		jobResult, err = ClientSet.BatchV1().Jobs("default").Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil {
			klog.Error("Create job error: ", err)
			return err
		}
		klog.Info("First delete job and create job complete, job result: ", jobResult)

		return nil
	}
	if err != nil {
		klog.Error("Create job error: ", err)
		return err
	}
	klog.Info("Create job complete, job result: ", jobResult)
	return nil
}

func getJobTaskName(taskName string, t string) string {
	var res string
	if t == common.ImageType {
		res = fmt.Sprintf("inspect-image-task-%v", taskName)
	} else if t == common.ScriptType {
		res = fmt.Sprintf("inspect-script-task-%v", taskName)
	}
	return res
}

// getContainerEnv 放入容器执行过程会用到的环境变量
func getContainerEnv(task *inspectv1alpha1.Task) []v1.EnvVar {
	eList := make([]v1.EnvVar, 0)
	if task.Type == common.ScriptType {
		e := v1.EnvVar{
			Name:  "script",
			Value: common.EncodeScript(task.Script),
		}
		eList = append(eList, e)
	}

	e1 := v1.EnvVar{
		Name:  "taskName",
		Value: task.TaskName,
	}
	e2 := v1.EnvVar{
		Name:  "type",
		Value: task.Type,
	}
	// FIXME: 不要写死，改为动态配置
	e3 := v1.EnvVar{
		Name:  "message-operator-url",
		Value: "http://42.193.17.123:31130/v1/send",
	}
	e4 := v1.EnvVar{
		Name:  "script_location",
		Value: task.ScriptLocation,
	}

	// 如果是远端节点，把user password ip 注入环境变量
	if task.ScriptLocation == "remote" {
		for _, v := range task.RemoteIps {
			eUser := v1.EnvVar{
				Name:  "user",
				Value: v.User,
			}
			ePassword := v1.EnvVar{
				Name:  "password",
				Value: v.Password,
			}
			eIp := v1.EnvVar{
				Name:  "ip",
				Value: v.Ip,
			}
			eList = append(eList, eUser)
			eList = append(eList, ePassword)
			eList = append(eList, eIp)
		}
	}

	eList = append(eList, e1)
	eList = append(eList, e2)
	eList = append(eList, e3)
	eList = append(eList, e4)
	return eList
}

func setJobTaskName(taskName string, t string) string {
	var res string
	if t == common.ImageType {
		res = fmt.Sprintf("inspect-image-task-%v", taskName)
	} else if t == common.ScriptType {
		res = fmt.Sprintf("inspect-script-task-%v", taskName)
	}
	return res
}

func jobSpec(task *inspectv1alpha1.Task, image string) *batchv1.Job {
	klog.Info("Create k8s job params: taskName= ", task.TaskName)
	// init
	taskName := setJobTaskName(task.TaskName, task.Type)
	// FIXME: 使用时间戳会有定时任务查不到的问题
	//taskName := fmt.Sprintf("inspect-manual-task-%v-%d", task.TaskName, time.Now().Unix())
	labels := map[string]string{"taskType": "inspect", "app": taskName}
	return &batchv1.Job{
		// metadata
		ObjectMeta: metav1.ObjectMeta{
			Name:   taskName,
			Labels: labels,
		}, Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					// containers
					Containers: []v1.Container{{
						Name:            "inspect-container",
						Image:           image,
						ImagePullPolicy: v1.PullIfNotPresent,
						Env:             getContainerEnv(task),
					}},
					// restart policy
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: common.DefaultServiceAccount,
				},
			},
		},
	}
}

// DeleteJob
// @see https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/job-v1/#delete-delete-a-job
func DeleteJob(tasks []*inspectv1alpha1.Task) error {

	ids := make([]string, len(tasks))
	for i, task := range tasks {
		ids[i] = task.TaskName
	}
	klog.Info("Delete k8s job, task ids: ", ids)
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("taskId in (%s)", strings.Join(ids, ",")),
	}
	foreground := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &foreground}
	return ClientSet.BatchV1().Jobs("default").DeleteCollection(context.TODO(), deleteOptions, listOptions)
}
