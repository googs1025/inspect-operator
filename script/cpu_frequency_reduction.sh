#!/bin/bash
# 检查是否有CPU降频


# 获取 CPU 的最大频率
max_freq=$(cat /sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq)

# 获取 CPU 当前的频率
current_freq=$(cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq)

# 将频率值转换为 MHz
max_freq=$((max_freq / 1000))
current_freq=$((current_freq / 1000))

# 检查是否存在 CPU 降频
if [ "$current_freq" -lt "$max_freq" ]; then
  echo "存在 CPU 降频！"
  echo "最大频率: ${max_freq} MHz"
  echo "当前频率: ${current_freq} MHz"
else
  echo "没有 CPU 降频。"
fi