#!/bin/bash
# 检查是否有磁盘使用率大于80%

# 获取根文件系统的使用率
usage=$(df -h | awk '$NF=="/"{print $5}' | sed 's/%//')

# 检查使用率是否大于 80%
if [ "$usage" -gt 80 ]; then
  echo "磁盘使用率超过 80%！"
else
  echo "磁盘使用率正常。"
fi