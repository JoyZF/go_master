# K8S 实战
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
- pod.yaml的image直接写docker registry 的镜像地址，报ErrPullImage.
修改minikube启动命令  指定registry.
minikube start --memory 2048mb --cpus 2 \
  --cache-images=true \
  --driver=docker \
  --image-mirror-country=cn \
  --insecure-registry='127.0.0.1:5000' \
  --registry-mirror="https://registry.docker-cn.com,https://docker.mirrors.ustc.edu.cn" \
  --service-cluster-ip-range='10.10.0.0/24'


