# 将博客使用minikube部署遇到的一些问题
需求：将bbpp.onilne 使用k8s 部署。
## 需求拆分
### 运行环境
- minikube
### 遇到的问题
#### Minikube start 报 "Exiting due to DRV_AS_ROOT: The “docker“ driver should not be used with root privileges."
解决办法：
```shell
minikube start --force --driver=docker
```
#### alias kubectl="minikube kubectl --" 每次开一个新的tel都要重新设置一次
[](https://www.python100.com/html/80513.html)

#### minikube dashboard外网访问
```shell
kubectl proxy --port=80 --address=0.0.0.0 --accept-hosts=^.*
```

#### kubectl port-forward service/hello-minikube 7080:8080  
通过端口转发映射本地端口到指定的应用端口，从而访问集群中的应用程序(Pod).
但是通过这种方式，只能访问到本地的端口。
如果要让外网也能访问到的话，需要带上--address 0.0.0.0 或者指定访问IP。
```shell
kubectl port-forward --address 0.0.0.0 service/hello-minikube 7080:8080
```

#### kubectl 后台运行
用nohup后台运行时会报 nohup: 无法运行命令"kubectl": 没有那个文件或目录
那就另辟蹊径，使用screen命令：screen命令可以创建一个虚拟终端窗口，并在其中运行kubectl命令。这样可以随时在后台访问和管理kubectl会话，而不受SSH会话的限制。

```shell
sudo yum install screen

screen -S kubectl_session

kubectl [command]

```

按下Ctrl + A，然后按下D，将screen会话切换到后台。

要重新连接到已经在后台运行的screen会话，可以执行以下命令.
```shell
screen -r kubectl_session

```


### 启动一个Nginx Pod 
### build 一个docker image
```shell
// 以nginx为例
1、 docker pull nginx 
2、docker images 查看是否pull成功
3、对改nginx做一些改动
4、docker tag yourdomain/nginx:latest
5、docker push yourdomain/nginx:latest
6、curl https://yourdomain/v2/_catalog
返回：{"repositories":["myubuntu","nginx"]} 即上传成功
7、docker pull yourdomain/nginx:latest 如果pull即成功
```

[Docker registry](https://docs.docker.com/registry/)





### docker 命令记录
- docker commit -a "作者" -m "提交信息" 容器ID 仓库名:TAG 从容器创建一个新的镜像
- docker tag [OPTIONS] IMAGE[:TAG] [REGISTRYHOST/][USERNAME/]NAME[:TAG] 标记本地镜像，将其归入某一仓库。
- 


### bbpp.online 部署到k8s环境
- 把hugo的public目录打包进nginx docker image
- create nginx pod
- create nginx service
- create ingress
- 宿主机的nginx 转发到ingress

#### 遇到的问题
##### pod.yaml的image直接写docker registry 的镜像地址，报ErrPullImage.
修改minikube启动命令  指定registry.
minikube start --memory 2048mb --cpus 2 \
  --cache-images=true \
  --driver=docker \
  --image-mirror-country=cn \
  --insecure-registry='docker.bbpp.online' \
  --registry-mirror="https://registry.docker-cn.com,https://docker.mirrors.ustc.edu.cn" \
  --service-cluster-ip-range='10.10.0.0/24'

[Pod 一直处于 ImagePullBackOff 状态](https://cloud.tencent.com/document/product/457/42947)

##### pod 一直 CrashLoopBackOff
经排查，这次CrashLoopBackOff 是因为yaml中声明了一个exec： /bin/echo 实际上容器中是没有echo 命令的，所以一直报错，去掉多余的声明即可。

其他CrashLoopBackOff排查思路：
- 系统发生OOM 以看到 Pod 中容器退出状态码是 137，表示被 SIGKILL 信号杀死，同时内核会报错: Out of memory: Kill process ...。大概率是节点上部署了其它非 K8S 管理的进程消耗了比较多的内存，或者 kubelet 的 --kube-reserved 和 --system-reserved 配的比较小，没有预留足够的空间给其它非容器进程，节点上所有 Pod 的实际内存占用总量不会超过 /sys/fs/cgroup/memory/kubepods 这里 cgroup 的限制，这个限制等于 capacity - "kube-reserved" - "system-reserved"，如果预留空间设置合理，节点上其它非容器进程（kubelet, dockerd, kube-proxy, sshd 等) 内存占用没有超过 kubelet 配置的预留空间是不会发生系统 OOM 的，可以根据实际需求做合理的调整。
- cgroup OOM 如果是 cgrou OOM 杀掉的进程，从 Pod 事件的下 Reason 可以看到是 OOMKilled，说明容器实际占用的内存超过 limit 了，同时内核日志会报: ``。 可以根据需求调整下 limit。
- 节点内存碎片化 如果节点上内存碎片化严重，缺少大页内存，会导致即使总的剩余内存较多，但还是会申请内存失败，参考 处理实践: 内存碎片化
详细内容可参考腾讯云的这篇文章：[Pod 处于 CrashLoopBackOff 状态](https://cloud.tencent.com/document/product/457/43130)

### Ingress Controller , Ingress Class ,Ingress , Service , Pod 的相互联系
![](https://static001.geekbang.org/resource/image/bb/14/bb7a911e10c103fb839e01438e184914.jpg?wh=1920x736)
