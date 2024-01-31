package cronjob

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/myoperator/inspectoperator/pkg/common"
	"github.com/myoperator/inspectoperator/pkg/sysconfig"
	. "github.com/myoperator/inspectoperator/test/workload"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// CreateCronJob
// @see https://v1-20.docs.kubernetes.io/docs/reference/kubernetes-api/workload-resources/cron-job-v1beta1/#create-create-a-cronjob
func CreateCronJob(task *sysconfig.Task, image string) error {

	cronjob := cronJobSpec(task, image)

	cronjobBytes, _ := json.Marshal(cronjob)
	fmt.Printf("Create cronjob with params: %s\n", string(cronjobBytes))

	// create cronjob
	jobResult, err := ClientSet.BatchV1beta1().CronJobs("namespace").Create(context.TODO(), cronjob, metav1.CreateOptions{})
	fmt.Printf("Create cronjob complete, job result: %v\n", jobResult)
	return err
}

// UpdateCronJob
// @see https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/cron-job-v1/#update-replace-the-specified-cronjob
func UpdateCronJob(task *sysconfig.Task, image string) error {

	cronjob := cronJobSpec(task, image)

	cronjobBytes, _ := json.Marshal(cronjob)
	fmt.Printf("Update cronjob with params: %s\n", string(cronjobBytes))

	// update cronjob
	jobResult, err := ClientSet.BatchV1beta1().CronJobs("namespace").Update(context.TODO(), cronjob, metav1.UpdateOptions{})
	fmt.Printf("Update cronjob complete, job result: %v\n", jobResult)
	return err
}

func cronJobSpec(task *sysconfig.Task, image string) *v1beta1.CronJob {
	// init
	taskName := cronJobName(task.TaskName)
	labels := map[string]string{"taskType": "inspect", "app": taskName}

	schedule := common.ChangeCronExpressionTimeZone("ttttttt")
	cronjob := &v1beta1.CronJob{
		// metadata
		ObjectMeta: metav1.ObjectMeta{
			Name:   taskName,
			Labels: labels,
		},
		// spec
		Spec: v1beta1.CronJobSpec{
			Schedule:                schedule,
			ConcurrencyPolicy:       v1beta1.ForbidConcurrent,
			StartingDeadlineSeconds: &common.K8SStartingDeadlineSeconds,
			JobTemplate: v1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					ActiveDeadlineSeconds:   &common.K8SJobTimeoutSeconds,
					BackoffLimit:            &common.K8SJobBackoffLimit,
					TTLSecondsAfterFinished: &common.K8SCronJobTTLSeconds,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: labels,
						},
						Spec: v1.PodSpec{
							// containers
							Containers: []v1.Container{{
								Name:            "namespace",
								Image:           image,
								ImagePullPolicy: v1.PullIfNotPresent,
							}},
							// restart policy
							RestartPolicy:      v1.RestartPolicyNever,
							ServiceAccountName: common.DefaultServiceAccount,
						},
					},
				},
			},
			SuccessfulJobsHistoryLimit: &common.K8SJobHistoryLimit,
			FailedJobsHistoryLimit:     &common.K8SJobFailedHistoryLimit,
		},
	}
	return cronjob
}

func cronJobName(taskName string) string {
	return fmt.Sprintf("inspect-schedule-task-%v", taskName)
}

// DeleteCronJob
// @see https://v1-20.docs.kubernetes.io/docs/reference/kubernetes-api/workload-resources/cron-job-v1beta1/#delete-delete-a-cronjob
func DeleteCronJob(tasks ...*sysconfig.Task) error {

	names := make([]string, len(tasks))
	for i, task := range tasks {
		names[i] = cronJobName(task.TaskName)
	}

	fmt.Printf("Delete k8s cronjob, task names: %v\n", tasks)
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app in (%s)", strings.Join(names, ",")),
	}
	foreground := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &foreground}
	return ClientSet.BatchV1beta1().CronJobs("namespace").DeleteCollection(context.TODO(), deleteOptions, listOptions)
}
