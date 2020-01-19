# OneBD

## 设计思路


## 核心对象

- application

> 整体应用的全局配置和管理

- handler 

> 请求的周期管理

- handlerPool

> handler缓存池 避免每次去创建

- router

> 路由 根据请求路径匹配正确的handler去处理

- meta 

> 辅助handler处理request和response 


## TODO:

- goroutine pool

- hook

- log

- error

- cache

- distribute

- websocket

- MQ

- session

- auth

- 


## 注意

- 本项目多次使用对象缓存和复用技术，谨慎使用衍生go程读取或修改原go程数据