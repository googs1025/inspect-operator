#!/bin/bash
# 检查内存的空闲率是否大于80%

# 从 /proc/meminfo 文件中提取内存信息
mem_info=$(grep -E 'MemTotal|MemAvailable' /proc/meminfo)

# 提取内存总量和可用内存
total_mem=$(echo "$mem_info" | awk '{print $2}')
available_mem=$(echo "$mem_info" | awk '{print $2}')

# 计算内存使用率
usage_percentage=$(( (total_mem - available_mem) * 100 / total_mem ))

# 检查使用率是否大于 80%
if [ "$usage_percentage" -gt 80 ]; then
  echo "内存使用率超过 80%！"
else
  echo "内存使用率正常。"
fi