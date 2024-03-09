#!/bin/bash
# 检查是否有僵尸进程
#export TERM=xterm
#zombieCount=$(top -bn 1 |grep 'Tasks' | awk '{print $10}')
#if [ $zombieCount -gt 0 ]
#then
#    echo caseName:无僵尸进程, caseDesc:, result:fail, resultDesc:有${zombieCount}个僵尸进程
#else
#    echo caseName:无僵尸进程, caseDesc:, result:success, resultDesc:无僵尸进程
#fi


# 使用 ps 命令查找僵尸进程
zombie_processes=$(ps -eo stat,pid | grep -w Z)

# 检查是否存在僵尸进程
if [ -n "$zombie_processes" ]; then
  echo "存在僵尸进程！"
  echo "僵尸进程列表："
  echo "$zombie_processes"
else
  echo "没有僵尸进程。"
fi