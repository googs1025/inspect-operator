package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	inspectv1alpha1 "github.com/myoperator/inspectoperator/pkg/apis/inspect/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type InspectController struct {
	client client.Client
	Scheme *runtime.Scheme
	event  record.EventRecorder
	log    logr.Logger
}

func NewInspectController(client client.Client, log logr.Logger, scheme *runtime.Scheme, event record.EventRecorder) *InspectController {
	return &InspectController{
		client: client,
		log:    log,
		event:  event,
		Scheme: scheme,
	}
}

// Reconcile 调协 loop
func (r *InspectController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	klog.Info("start inspect Reconcile..........")

	// load JobFlow by namespace
	inspect := &inspectv1alpha1.Inspect{}
	time.Sleep(time.Second)
	err := r.client.Get(ctx, req.NamespacedName, inspect)
	if err != nil {
		// If no instance is found, it will be returned directly
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("not found jobFlow : %v", req.Name))
			return reconcile.Result{}, nil
		}
		klog.Error(err, err.Error())
		r.event.Eventf(inspect, corev1.EventTypeWarning, "Created", err.Error())
		return reconcile.Result{}, err
	}

	if inspect.Status.Status == inspectv1alpha1.Failed {
		return reconcile.Result{}, nil
	}

	// FIXME: 处理 Finalizer 字段
	// 考虑是否要在 inspect status state 为 Running 时 不能删除？

	// deploy job by dependence order.
	if err = r.deployInspect(ctx, *inspect); err != nil {
		klog.Error("deployJob error: ", err)
		r.event.Eventf(inspect, corev1.EventTypeWarning, "Failed", err.Error())
		// 如果是 执行 job 任务出错，跳转
		return reconcile.Result{RequeueAfter: time.Second * 60}, err
	}

	// update status
	// 修改 job 狀態，list 出所有相關的 job ，並查看其狀態，並存在 status 中
	if err = r.updateJobFlowStatus(ctx, inspect); err != nil {
		klog.Error("update inspect status error: ", err)
		r.event.Eventf(inspect, corev1.EventTypeWarning, "Failed", err.Error())
		return reconcile.Result{}, err
	}
	klog.Info("end inspect Reconcile........")

	return reconcile.Result{}, nil
}
