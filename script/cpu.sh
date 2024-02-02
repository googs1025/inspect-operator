#!/bin/bash
# 检查cpu的空闲率是否小于20%

# 使用 mpstat 命令获取 CPU 使用率信息（每一秒采样一次，持续两秒）
cpu_usage=$(mpstat 1 2 | awk '/Average:/ {print $NF}')

# 检查 CPU 使用率是否小于 80%
if (( $(echo "$cpu_usage < 80" | bc -l) )); then
  echo "CPU 使用率小于 80%！"
  echo "当前使用率: $cpu_usage%"
else
  echo "CPU 使用率正常。"
fi