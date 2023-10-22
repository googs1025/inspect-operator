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
    - task:
        task_name: task1
        type: script             # type 字段：可填写脚本或镜像 image script 两种
        source: test.sh          # source 字段：可执行 是 bash 脚本 py 脚本或是镜像。需要把东西放入 ./script 中
        script_location: local   # script_location 字段(目前未支持此功能) 可选填 local remote all 三种，分别对应 本地节点 远端节点 全部节点
        # 选取 local 本地节点 就不需要再填写远端 ip 地址
        # 远端要执行的目标 node
        # script 字段：可填写 bash 脚本内容，controller 默认如果有 script 字段，优先执行自定义脚本内容，"不执行" source 字段脚本内容
        script: |
          # 检查是否有CPU降频
          count=0
          for cpuHz in $(cat /proc/cpuinfo | grep MHz | awk '{print $4}')
          do
              if [ `echo "$cpuHz < 2000.0" |bc` -eq 1 ]
              then
                  count=`expr $count + 1`
              fi
          done
          if [ $count -gt 0 ]
          then
              echo caseName:无CPU降频, caseDesc:, result:fail, resultDesc:有${count}个CPU的频率低于2000MHz, 可能发生降频
          else
              echo caseName:无CPU降频, caseDesc:, result:success, resultDesc:CPU频率都大于2000MHz, 无降频
          fi
    - task:
        task_name: task2
        type: image
        source: try:latest  # source 字段：镜像名称
        restart: true       # 用于标示是否重新执行。 如
    - task:
        task_name: task3
        type: script             # type 字段：可填写脚本或镜像 image script 两种
        script_location: remote   # script_location 字段(目前未支持此功能) 可选填 local remote all 三种，分别对应 本地节点 远端节点 全部节点
        # 远端要执行的目标 node 的信息：user password ip 等
        remote_ips:
          - user: "root"
            password: "xxxxxx"
            ip: "xxxxxx"
        script: |
          # 检查是否有CPU降频
          count=0
          for cpuHz in $(cat /proc/cpuinfo | grep MHz | awk '{print $4}')
          do
              if [ `echo "$cpuHz < 2000.0" |bc` -eq 1 ]
              then
                  count=`expr $count + 1`
              fi
          done
          if [ $count -gt 0 ]
          then
              echo caseName:无CPU降频, caseDesc:, result:fail, resultDesc:有${count}个CPU的频率低于2000MHz, 可能发生降频
          else
              echo caseName:无CPU降频, caseDesc:, result:success, resultDesc:CPU频率都大于2000MHz, 无降频
          fi
```
- 创建 script 巡检任务需要的镜像(内置镜像)
```bash
[root@VM-0-16-centos inspectoperator]# cd scriptimage/
[root@VM-0-16-centos scriptimage]# docker build -t inspect-operator/script-engine:v1 .
Sending build context to Docker daemon   29.7kB
Step 1/18 : FROM golang:1.18.7-alpine3.15 as builder
 ---> 33c97f935029
Step 2/18 : WORKDIR /app
 ---> Using cache
 ---> 8cc02fd966d4
Step 3/18 : COPY go.mod go.mod
 ---> Using cache
 ---> 6e1bcad7a69d
Step 4/18 : COPY go.sum go.sum
 ---> Using cache
```
- 创建自定义镜像(以 try:v1 为例) [参考 try 目录](./test/try)
```bash
[root@VM-0-16-centos try]# pwd
/root/inspectoperator/test/try
[root@VM-0-16-centos try]# docker build -t try:v1 .
Sending build context to Docker daemon  1.865MB
Step 1/8 : FROM golang:1.17
 ---> 742df529b073
```

### RoadMap
1. 实现远端局点执行脚本的能力
2. 优化发送结果回调通知的能力
3. 实现下发 cronjob 定时巡检能力
