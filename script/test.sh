#!/bin/bash
# 检查 k8s 中不正常状态 pod
# 使用 kubectl 命令获取所有 Pod 的状态信息
pod_status=$(kubectl get pods --all-namespaces -o jsonpath='{range .items[*]}{@.metadata.name}{"\t"}{@.metadata.namespace}{"\t"}{@.status.phase}{"\n"}{end}')

# 检查所有 Pod 的状态是否为 "Running"
echo "$pod_status" | while read -r name namespace status; do
  if [[ "$status" != "Running" ]]; then
    echo "不正常的 Pod:"
    echo "名称: $name"
    echo "命名空间: $namespace"
    echo "状态: $status"
    echo "----------------------"
  fi
done