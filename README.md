# OneBD

Auto generate code for backend


## model

- 文件夹 资源作用域

- 文件名 资源

- 文件内同名结构体 主资源

- 文件内不同结构体 主资源附属资源，一对多关系


## 路由

### 匹配规则

/path/:param1/:param2/*param3

### 错误处理

404 根路由触发
500 父路由递归触发一次
