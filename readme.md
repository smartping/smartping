#SmartPing

SmartPing为一个各机器(点)间间互PING检测工具，支持互PING，单向PING，绘制拓扑及报警功能。

## 功能 ##

- 单向PING，反向Ping绘图
- 互PING间机器的状态拓扑
- 自定义延迟、丢包阈值报警

## 设计思路 ##

本系统设计为无中心化原则，所有的数据均存储自身点中，默认每个Ping目标点的数据循环保留1个月时间，由自身点的数据绘制 **出PING包** 的状态，由各其他点的数据绘制 **进PING包** 的状态，从任意一点查询数据均会通过Ajax请求关联点的API接口获取其他点数据组装全部数据，绘制 出Ping曲线图，进Ping曲线图，网络互Ping拓扑图。并可以设置阈值进行报警，方便对网络质量的监控。

- 程序架构
- 数据结构


## 其他资料 ##

- 官网： http://smartping.org
- 文档： https://docs.smartping.org
- - 下载安装：https://docs.smartping.org/install.html
- - API文档：https://docs.smartping.org/api.html