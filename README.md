
# 概要

延伸自 doraemon的 rule-engine，用于与 promethues-alertmanager 结合

目的是，为 AlertManager 加入动态规则

## 功能

自带DB表，定时(1 min)从DB中读取告警规则，到对应数据源(即 Prometheus 服务)执行查询，将触发的告警，发送给配置文件指定的 AlertManager 服务

NOTE：
    向 AlertManager 发送的 alert 消息，暂时还是 doraemon 协议格式的，应该需要少量改造

## 其他
有两个辅助项目：
* [mysql4Prom](https://github.com/huangwei2013/mysql4prom)：用于解析 prometheus 规则文件，并导入DB
* [mysql4PromUI](https://github.com/huangwei2013/mysql4promUI)：用于 DB 对应的简单管理界面
