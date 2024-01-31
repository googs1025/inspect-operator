## inspect-operator 简易型集群内巡检中心
![](https://github.com/Operator-Learning-Playground/inspect-operator/blob/main/image/%E6%B5%81%E7%A8%8B%E5%9B%BE%20(1).jpg?raw=true)
### 项目思路与设计
设计背景：本项目基于 k8s 的扩展功能，实现 Inspect 的自定义资源，实现一个集群内的执行bash脚本或是自定义镜像的 controller 应用。调用方可在 cluster 中部署与启动相关配置即可使用。
思路：当应用启动后，会启动一个 controller，controller 会监听所需的资源，并执行相应的业务逻辑(如：执行巡检脚本或镜像，再使用集群内的消息中心进行通知)。

### 项目功能
1. 支持对集群内使用 job 执行用户自定义镜像内容功能(用户必须完成 image 开发部分，可参考 test/try 目录)。
2. 支持对本地节点执行 bash 脚本功能，其中提供内置巡检 bash 脚本或用户可自定义 bash 脚本内容。
3. 提供发送结果通知功能(使用集群内消息中心 operator 实现)。

- 自定义资源如下所示
```yaml
apiVersion: api.practice.com/v1alpha1
kind: Inspect
metadata:
  name: myinspect
spec:
  tasks:
    - task_name: task1
      type: script             # type 字段：可填写脚本或镜像 image script 两种
      script_location: local   # script_location 字段(目前未支持此功能) 可选填 local remote all 三种，分别对应 本地节点 远端节点 全部节点
      # script字段：可填写 bash 脚本内容
      script: |
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
    - task_name: task2
      type: image
      source: try:latest  # source 字段：镜像名称
    - task_name: task3
      type: script             # type字段：可填写脚本或镜像 image script 两种
      script_location: remote  # script_location字段：可选填 local remote，分别对应 本地节点 远端节点
      # 远端要执行的目标node的信息：user password ip 等
      remote_infos:
        - user: "root"
          password: "googs1025Aa"
          ip: "1.14.120.233"
      script: |
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
```
- 创建 script 巡检任务需要的镜像(内置镜像)
```bash
[root@VM-0-16-centos inspectoperator]# ./build_engine.sh
Sending build context to Docker daemon  35.33kB
Step 1/18 : FROM golang:1.18.7-alpine3.15 as builder
 ---> 33c97f935029
Step 2/18 : WORKDIR /app
 ---> Using cache
 ---> 8cc02fd966d4
Step 3/18 : COPY go.mod go.mod
 ---> Using cache
 ---> 4b7680d53e60
```

- 创建自定义镜像(以 try:v1 为例) [参考 try 目录](example/try)
```bash
[root@VM-0-16-centos try]# pwd
/root/inspectoperator/example/try
[root@VM-0-16-centos try]# docker build -t try:v1 .
Sending build context to Docker daemon  1.865MB
Step 1/8 : FROM golang:1.17
 ---> 742df529b073
```

### RoadMap
1. 实现远端局点执行脚本的能力
2. 优化发送结果回调通知的能力
3. 实现下发 cronjob 定时巡检能力
