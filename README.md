# OneBD

[TOC]

## 设计思路


## 路由

核心函数为注册和子路由

- SubRouter(prefix string) Router
    - prefix 匹配规则
        1. 按照第一次注册时间优先匹配
        2. 形如 /abc/efd/type:varA/
        3. 前后 / 不影响匹配，会自动填充前缀和删去后缀
        4. 有效字符 :/A-Za-z0-9
        5. type 支持类型: '', str, int, float
            - '' 匹配任意字符，不包含 /
            - str 严格匹配字符串， 必须包含A-Za-z
            - int 严格匹配数字
            - float 严格匹配浮点数 必须包含 '.'
            - 考虑是否添加类型注册功能，提供全功能正则匹配
        6. 路由覆盖, 匹配规则前者会覆盖后者，后者要生效需提前注册顺序
            - /path/:varA > /path/str:varA > /path/abc
            - /path/:varA > /path/int:varA > /path/0
- Set(prefix string, fc func() Handler, allowedMethods ...rfc.Method)
    - prefix 匹配规则同上
    - fc 生成新handler的方法，用于构建handler池，避免每次生成
    - allowedMethods 运行通过的http方法, 缺省则全部允许

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