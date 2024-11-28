# CRUD middleware


### json tag

method: "*Opt1@get@/:id,Opt2"

Opt1: 代表操作名称， 默认为get,post,delete,patch,put,list
*代表该参数在该操作下为可选参数
@get,@post,@delete,@patch,@put,@list: 代表该参数在该操作下的请求方法
@/urlsuffix: 代表该操作的url后缀,必须带上"/"

parse:"json@aliasName"
第一个为解析方式，目前支持json,form,query,header,path, 默认为json
第二个为别名，用于在请求中获取该参数的值, 存储在args中的key仍然为参数snake_name名或者jsontag 标记的名字
