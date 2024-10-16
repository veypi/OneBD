# OneBD

Auto generate code for backend


## model

### 模型创建
- 文件夹 资源隔离域

2级资源制
- 文件名 资源

- 文件内同名结构体 主资源

- 文件内不同结构体 主资源附属资源，一对多关系

### 模型字段

parse: path, query, header, json/form
parse:path@varOtherName
json和form在一个模型中只能有一类,不用显性标注，根据request请求自动分析, json为默认解析方法,其余类型参数会再次基础上覆盖

id 推荐使用uuid, 请求来源一般来自路径
通用参数如created_at,skip, page等，来源一般来自query
资源特异性参数放到body里，一般使用json


## 路由

### 匹配规则

/path/:param1/:param2/*param3

### 错误处理

404 根路由触发
500 父路由递归触发一次
