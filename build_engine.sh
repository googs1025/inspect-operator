#!/bin/bash

# 进入到 inspect_script_engine 目录
cd inspect_script_engine

# 构建镜像
docker build -t inspect-operator/script-engine:v1 .

# 打印构建完成的信息
echo "镜像构建完成：inspect-operator/script-engine:v1"