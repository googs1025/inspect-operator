#!/bin/bash

# 获取内存使用率的函数
get_memory_usage() {
    # 使用free命令获取内存信息，第二行的第三列为已使用的内存值
    total_memory=$(free | awk 'NR==2 {print $2}')
    used_memory=$(free | awk 'NR==2 {print $3}')
    # 计算内存使用率
    memory_usage=$(awk "BEGIN {printf \"%.2f\", $used_memory / $total_memory * 100}")
    echo $memory_usage
}

# 检查内存使用率是否超过80%
memory_usage=$(get_memory_usage)
echo "当前内存使用率: $memory_usage%"

if (( $(echo "$memory_usage > 80" | bc -l) )); then
    echo "内存使用率超过80%！"
    # 在此处可以添加额外的操作，如发送警报通知等。
fi