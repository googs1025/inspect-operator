package main

import (
	inspectv1alpha1 "github.com/myoperator/inspectoperator/pkg/apis/inspect/v1alpha1"
	"github.com/myoperator/inspectoperator/pkg/controller"
	"github.com/myoperator/inspectoperator/pkg/k8sconfig"
	batchv1 "k8s.io/api/batch/v1"
	_ "k8s.io/code-generator"
	"k8s.io/klog/v2"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

/*
	manager 主要用来管理Controller Admission Webhook 包括：
	访问资源对象的client cache scheme 并提供依赖注入机制 优雅关闭机制

	operator = crd + controller + webhook
*/

func main() {

	logf.SetLogger(zap.New())
	var d time.Duration = 0
	// 1. 管理器初始化
	mgr, err := manager.New(k8sconfig.K8sRestConfig(), manager.Options{
		Logger:     logf.Log.WithName("inspect-operator"),
		SyncPeriod: &d, // resync不设置触发
	})
	if err != nil {
		mgr.GetLogger().Error(err, "unable to set up manager")
		os.Exit(1)
	}

	// 2. ++ 注册进入序列化表
	err = inspectv1alpha1.SchemeBuilder.AddToScheme(mgr.GetScheme())
	if err != nil {
		klog.Error(err, "unable add schema")
		os.Exit(1)
	}

	// 3. 控制器相关
	inspectCtl := controller.NewInspectController(mgr.GetClient(), mgr.GetLogger(),
		mgr.GetScheme(), mgr.GetEventRecorderFor("inspect-operator"))

	err = builder.ControllerManagedBy(mgr).
		For(&inspectv1alpha1.Inspect{}).
		Watches(&source.Kind{Type: &batchv1.Job{}},
			handler.Funcs{
				UpdateFunc: inspectCtl.OnUpdateJobHandlerByInspect,
				DeleteFunc: inspectCtl.OnDeleteJobHandlerByInspect,
			},
		).Watches(&source.Kind{Type: &batchv1.CronJob{}},
		handler.Funcs{
			UpdateFunc: inspectCtl.OnUpdateCronJobHandlerByInspect,
			DeleteFunc: inspectCtl.OnDeleteCronHandlerByInspect,
		},
	).Complete(inspectCtl)

	errC := make(chan error)

	// 5. 启动controller管理器
	go func() {
		klog.Info("controller start!! ")
		if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
			errC <- err
		}
	}()

	// 这里会阻塞，两种常驻进程可以使用这个方法
	getError := <-errC
	log.Println(getError.Error())

}
