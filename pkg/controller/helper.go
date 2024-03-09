package controller

import (
	"context"
	"fmt"
	inspectv1alpha1 "github.com/myoperator/inspectoperator/pkg/apis/inspect/v1alpha1"
	"github.com/myoperator/inspectoperator/pkg/common"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// deploy job by dependence order.
func (r *InspectController) deployInspect(ctx context.Context, inspect inspectv1alpha1.Inspect) error {
	// 启动 job

	for _, task := range inspect.Spec.Tasks {

		jobName := getJobName(inspect.Name, task.TaskName)
		namespacedNameJob := types.NamespacedName{
			Namespace: inspect.Namespace,
			Name:      jobName,
		}
		// job 对象
		var job client.Object
		// TODO: 区分 job 跟 cronjob
		if inspect.Spec.Type == "" {
			inspect.Spec.Type = inspectv1alpha1.JobsType
		}
		if inspect.Spec.Type == inspectv1alpha1.JobsType {
			job = prepareJob(&inspect, &task, jobName)
		} else if inspect.Spec.Type == inspectv1alpha1.CronJobsType {
			job = prepareCronJob(&inspect, &task, jobName)
		}

		// 如果没拿到这个 job
		if err := r.client.Get(ctx, namespacedNameJob, job); err != nil {
			if errors.IsNotFound(err) {
				// 如果沒有，直接創建，如果有，則要判斷 Dependencies 中的 job 是否已經成功
				if err = r.client.Create(ctx, job); err != nil {
					if errors.IsAlreadyExists(err) {
						break
					}
					return err
				}
				r.event.Eventf(&inspect, v1.EventTypeNormal, "Created", fmt.Sprintf("create job named %v for next step", job.GetName()))
				continue
			}
			return err
		}
	}
	return nil
}

func getJobName(jobFlowName string, jobTemplateName string) string {
	return jobFlowName + "-" + jobTemplateName
}

// update status
func (r *InspectController) updateJobFlowStatus(ctx context.Context, inspect *inspectv1alpha1.Inspect) error {
	klog.Info(fmt.Sprintf("start to update inspect status! inspectName: %v, inspectNamespace: %v ", inspect.Name, inspect.Namespace))

	var inspectStatus *inspectv1alpha1.InspectStatus
	if inspect.Spec.Type == "" {
		inspect.Spec.Type = inspectv1alpha1.JobsType
	}
	// 处理 jobs or cronjobs
	if inspect.Spec.Type == inspectv1alpha1.JobsType {
		// 获取 job 列表
		allJobList := new(batchv1.JobList)
		err := r.client.List(ctx, allJobList)
		if err != nil {
			klog.Error("list error: ", err)
			return err
		}
		inspectStatus, err = r.getAllJobStatus(inspect, allJobList)
		if err != nil {
			return err
		}
	} else if inspect.Spec.Type == inspectv1alpha1.CronJobsType {
		// 获取 job 列表
		allCronJobList := new(batchv1.CronJobList)
		err := r.client.List(ctx, allCronJobList)
		if err != nil {
			klog.Error("list error: ", err)
			return err
		}
		inspectStatus, err = r.getAllCronJobStatus(inspect, allCronJobList)
		if err != nil {
			return err
		}
	}

	inspect.Status = *inspectStatus
	if inspectStatus.Status == inspectv1alpha1.Succeed || inspectStatus.Status == inspectv1alpha1.Failed {
		r.event.Eventf(inspect, v1.EventTypeNormal, inspectStatus.Status, fmt.Sprintf("finshed inspect named %s", inspect.Name))
	}
	if err := r.client.Status().Update(ctx, inspect); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

const (
	DefaultServiceAccount = "myinspect-sa"
	ScriptType            = "script"
	ImageType             = "image"
	// ScriptExecuteImage 内置运行脚本引擎
	ScriptExecuteImage = "inspect-operator/script-engine:v1"
	RemoteType         = "remote"
)

func prepareJob(inspect *inspectv1alpha1.Inspect, task *inspectv1alpha1.Task, jobName string) *batchv1.Job {
	// job 对象
	job := &batchv1.Job{}

	// 设置 ownerReferences
	job.OwnerReferences = append(job.OwnerReferences, metav1.OwnerReference{
		APIVersion: inspect.APIVersion,
		Kind:       inspect.Kind,
		Name:       inspect.Name,
		UID:        inspect.UID,
	})

	job.Name = jobName
	job.Namespace = inspect.Namespace
	job.ObjectMeta.Labels = map[string]string{
		"inspect-name": inspect.Name,
	}
	var cc int32
	job.Spec.BackoffLimit = &cc

	var imageName string
	switch task.Type {
	case ImageType:
		imageName = task.Source
	case ScriptType:
		imageName = ScriptExecuteImage
	}

	job.Spec = batchv1.JobSpec{
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"job-name":     jobName,
					"inspect-name": inspect.Name,
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:            "inspect-container",
						Image:           imageName,
						ImagePullPolicy: v1.PullIfNotPresent,
						Env:             getContainerEnv(inspect, task),
					},
				},
				RestartPolicy:      v1.RestartPolicyNever,
				ServiceAccountName: DefaultServiceAccount,
			},
		},
	}

	if inspect.Spec.GlobalParams.Annotations != nil {
		job.Annotations = inspect.Spec.GlobalParams.Annotations
		job.Spec.Template.Annotations = inspect.Spec.GlobalParams.Annotations
	}

	if inspect.Spec.GlobalParams.Labels != nil {
		job.Labels = inspect.Spec.GlobalParams.Labels
		job.Spec.Template.Labels = inspect.Spec.GlobalParams.Labels
	}

	if inspect.Spec.GlobalParams.NodeName != "" {
		job.Spec.Template.Spec.NodeName = inspect.Spec.GlobalParams.NodeName
	}

	return job
}

func prepareCronJob(inspect *inspectv1alpha1.Inspect, task *inspectv1alpha1.Task, jobName string) *batchv1.CronJob {
	// cronjob 对象
	cronjob := &batchv1.CronJob{}

	// 设置 ownerReferences
	cronjob.OwnerReferences = append(cronjob.OwnerReferences, metav1.OwnerReference{
		APIVersion: inspect.APIVersion,
		Kind:       inspect.Kind,
		Name:       inspect.Name,
		UID:        inspect.UID,
	})

	cronjob.Name = jobName
	cronjob.Namespace = inspect.Namespace
	cronjob.ObjectMeta.Labels = map[string]string{
		"inspect-name": inspect.Name,
	}

	var imageName string
	switch task.Type {
	case ImageType:
		imageName = task.Source
	case ScriptType:
		imageName = ScriptExecuteImage
	}
	if inspect.Spec.Schedule == "" {
		inspect.Spec.Schedule = "*/1 * * * *"
	}

	cronjob.Spec = batchv1.CronJobSpec{
		Schedule: inspect.Spec.Schedule,
		JobTemplate: batchv1.JobTemplateSpec{
			Spec: batchv1.JobSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"cronjob-name": jobName,
							"inspect-name": inspect.Name,
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:            "inspect-container",
								Image:           imageName,
								ImagePullPolicy: v1.PullIfNotPresent,
								Env:             getContainerEnv(inspect, task),
							},
						},
						RestartPolicy:      v1.RestartPolicyNever,
						ServiceAccountName: DefaultServiceAccount,
					},
				},
			},
		},
	}
	var cc int32
	cronjob.Spec.JobTemplate.Spec.BackoffLimit = &cc

	if inspect.Spec.GlobalParams.Annotations != nil {
		cronjob.Annotations = inspect.Spec.GlobalParams.Annotations
		cronjob.Spec.JobTemplate.Spec.Template.Annotations = inspect.Spec.GlobalParams.Annotations
	}

	if inspect.Spec.GlobalParams.Labels != nil {
		cronjob.Labels = inspect.Spec.GlobalParams.Labels
		cronjob.Spec.JobTemplate.Spec.Template.Labels = inspect.Spec.GlobalParams.Labels
	}

	if inspect.Spec.GlobalParams.NodeName != "" {
		cronjob.Spec.JobTemplate.Spec.Template.Spec.NodeName = inspect.Spec.GlobalParams.NodeName
	}

	return cronjob
}

// getContainerEnv 放入容器执行过程会用到的环境变量
func getContainerEnv(inspect *inspectv1alpha1.Inspect, task *inspectv1alpha1.Task) []v1.EnvVar {
	eList := make([]v1.EnvVar, 0)
	if task.Type == ScriptType {
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
	if task.ScriptLocation == RemoteType {
		for _, v := range task.RemoteInfos {
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

	for _, v := range inspect.Spec.GlobalParams.Env {
		eList = append(eList, v)
	}

	eList = append(eList, e1)
	eList = append(eList, e2)
	eList = append(eList, e3)
	eList = append(eList, e4)
	return eList
}

// getAllJobStatus 记录 Job Status
func (r *InspectController) getAllJobStatus(inspect *inspectv1alpha1.Inspect, allJobList *batchv1.JobList) (*inspectv1alpha1.InspectStatus, error) {
	// 过去掉只留 inspect 相关的 job
	jobListRes := make([]batchv1.Job, 0)
	for _, job := range allJobList.Items {
		for _, reference := range job.OwnerReferences {
			if reference.Kind == inspectv1alpha1.InspectKind && reference.Name == inspect.Name {
				jobListRes = append(jobListRes, job)
			}
		}
	}

	runningJobs := make([]string, 0)
	failedJobs := make([]string, 0)
	completedJobs := make([]string, 0)

	jobList := make([]string, 0)

	for _, task := range inspect.Spec.Tasks {
		jobList = append(jobList, getJobName(inspect.Name, task.TaskName))
	}

	inspectStatus := inspectv1alpha1.InspectStatus{
		JobResults: map[string]batchv1.JobStatus{},
	}

	for _, job := range jobListRes {
		a := fmt.Sprintf("%s/%s", job.Name, job.Namespace)

		inspectStatus.JobResults[a] = job.Status

		if job.Status.Succeeded == 1 {
			completedJobs = append(completedJobs, job.Name)
		} else if job.Status.Failed == 1 {
			failedJobs = append(failedJobs, job.Name)
		} else if job.Status.Active == 1 {
			runningJobs = append(runningJobs, job.Name)
		}
	}

	// 确认 jobFlow 狀態
	if inspect.DeletionTimestamp != nil {
		inspectStatus.Status = inspectv1alpha1.Terminating
	} else {
		if len(jobList) != len(completedJobs) {
			if len(failedJobs) > 0 {
				inspectStatus.Status = inspectv1alpha1.Failed
			} else if len(runningJobs) > 0 || len(completedJobs) > 0 {
				inspectStatus.Status = inspectv1alpha1.Running
			} else {
				inspectStatus.Status = inspectv1alpha1.Pending
			}
		} else {
			inspectStatus.Status = inspectv1alpha1.Succeed
		}
	}

	return &inspectStatus, nil
}

func (r *InspectController) getAllCronJobStatus(inspect *inspectv1alpha1.Inspect, allJobList *batchv1.CronJobList) (*inspectv1alpha1.InspectStatus, error) {
	// 过去掉只留 inspect 相关的 cronjob
	jobListRes := make([]batchv1.CronJob, 0)
	for _, job := range allJobList.Items {
		for _, reference := range job.OwnerReferences {
			if reference.Kind == inspectv1alpha1.InspectKind && reference.Name == inspect.Name {
				jobListRes = append(jobListRes, job)
			}
		}
	}

	runningJobs := make([]string, 0)
	failedJobs := make([]string, 0)
	completedJobs := make([]string, 0)

	jobList := make([]string, 0)

	for _, task := range inspect.Spec.Tasks {
		jobList = append(jobList, getJobName(inspect.Name, task.TaskName))
	}

	inspectStatus := inspectv1alpha1.InspectStatus{
		CronJobResults: map[string]batchv1.CronJobStatus{},
	}

	for _, job := range jobListRes {
		a := fmt.Sprintf("%s/%s", job.Name, job.Namespace)
		inspectStatus.CronJobResults[a] = job.Status

		// 根据 cronjob 字段 遍历寻找 job
		for _, v := range inspectStatus.CronJobResults {
			active := v
			for _, j := range active.Active {
				jobFromCronjob := &batchv1.Job{}
				namespacedNameJob := types.NamespacedName{
					Namespace: inspect.Namespace,
					Name:      j.Name,
				}
				// 如果没拿到这个 job
				err := r.client.Get(context.Background(), namespacedNameJob, jobFromCronjob)
				if errors.IsNotFound(err) {

					continue
				}
				if jobFromCronjob.Status.Succeeded == 1 {
					completedJobs = append(completedJobs, job.Name)
				} else if jobFromCronjob.Status.Failed == 1 {
					failedJobs = append(failedJobs, job.Name)
				} else if jobFromCronjob.Status.Active == 1 {
					runningJobs = append(runningJobs, job.Name)
				}

			}
		}

	}

	// 确认 jobFlow 狀態
	if inspect.DeletionTimestamp != nil {
		inspectStatus.Status = inspectv1alpha1.Terminating
	} else {
		if len(jobList) != len(completedJobs) {
			if len(failedJobs) > 0 {
				inspectStatus.Status = inspectv1alpha1.Failed
			} else if len(runningJobs) > 0 || len(completedJobs) > 0 {
				inspectStatus.Status = inspectv1alpha1.Running
			} else {
				// FIX: 这里可能会有些问题
				inspectStatus.Status = inspectv1alpha1.CronExecute
			}
		} else {
			inspectStatus.Status = inspectv1alpha1.CronExecute
		}
	}

	return &inspectStatus, nil
}

func (r *InspectController) OnUpdateCronJobHandlerByInspect(event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.ObjectNew.GetOwnerReferences() {
		if ref.Kind == inspectv1alpha1.InspectKind && ref.APIVersion == inspectv1alpha1.InspectApiVersion {
			// 重新放入 Reconcile 调协方法
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: ref.Name, Namespace: event.ObjectNew.GetNamespace(),
				},
			})
		}
	}
}

func (r *InspectController) OnDeleteCronHandlerByInspect(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.Object.GetOwnerReferences() {
		if ref.Kind == inspectv1alpha1.InspectKind && ref.APIVersion == inspectv1alpha1.InspectApiVersion {
			// 重新入列
			klog.Info("delete pod: ", event.Object.GetName(), event.Object.GetObjectKind())
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ref.Name,
					Namespace: event.Object.GetNamespace()}})
		}
	}
}
