package main

import (
	"context"
	. "github.com/myoperator/inspectoperator/test/workload"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func main() {
	// 测试取到job
	r, err := ClientSet.BatchV1().Jobs("default").Get(context.Background(), "res", v1.GetOptions{})

	if err != nil && errors.IsNotFound(err) {
		klog.Error("not found error ", err)
		return
	} else if err != nil {
		klog.Error("get job error ", err)
		return
	}
	klog.Info(r)

	// 取到message-operator的svc
	svc, err := ClientSet.CoreV1().Services("default").Get(context.Background(), "mymessage-svc", v1.GetOptions{})
	if err != nil {
		klog.Error(err)
	}
	klog.Info(svc.Status)
}
