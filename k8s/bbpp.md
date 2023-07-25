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