# grafana+prometheus
Prometheus 是一套开源的监控 + 预警 + 时间序列数据库的组合，现在越来越多的公司或组织开始采用 Prometheus，现在常见的 kubernetes 容器管理系统，也会搭配 Prometheus 来进行监控。
Prometheus 本身不具备收集监控数据功能，需要使用 http 接口来获取不同的 export 收集的数据，存储到时序数据库中。

## 搭建环境
- Prometheus： 普罗米修斯的主服务器,端口号9090
- NodeExporter：负责收集Host硬件信息和操作系统信息，端口号9100
- MySqld_Exporter：负责收集mysql数据信息收集，端口号9104
- cAdvisor：负责收集Host上运行的docker容器信息,端口号占用8080
- Grafana：负责展示普罗米修斯监控界面，端口号3000
- altermanager：等待接收prometheus发过来的告警信息，altermanager再发送给定义的收件人

## 安装grafana+prometheus
[使用 docker 搭建 grafana+prometheus 监控服务器资源](https://blog.csdn.net/Song_Lun/article/details/120666421)

[使用 docker 搭建 grafana+prometheus 监控数据库资源（贰）](https://blog.csdn.net/Song_Lun/article/details/120740732)

[使用 docker 搭建 grafana+prometheus 监控docker资源（叁）](https://blog.csdn.net/Song_Lun/article/details/120777812)

[使用 docker 搭建 grafana+prometheus+AlertManager 邮件报警（肆）](https://blog.csdn.net/Song_Lun/article/details/120748996)