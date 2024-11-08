//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

import axios, { AxiosError, type AxiosResponse } from 'axios';

// Be careful when using SSR for cross-request state pollution
// due to creating a Singleton instance here;
// If any client changes this (global) instance, it might be a
// good idea to move this instance creation inside of the
// "export default () => {}" function below (which runs individually
// for each client)

axios.defaults.withCredentials = true
const proxy = axios.create({
  withCredentials: true,
  baseURL: "/api/",
  headers: {
    'content-type': 'application/json;charset=UTF-8',
  },
});

export const token = {
  value: '',
  update: () => {
    return new Promise<string>((resolve) => { resolve(token.value) })
  },
  set_updator: (fn: () => Promise<string>) => {
    let locked = false
    token.update = () => {
      if (locked) {
        return new Promise<void>((resolve) => {
          setTimeout(() => {
            resolve();
          }, 100)
        }).then(() => {
          return token.update()
        })
      }
      locked = true
      return new Promise<string>((resolve, reject) => {
        if (token.value) {
          locked = false
          resolve(token.value)
          return
        }
        fn().then((e) => {
          token.value = e
          resolve(e)
        }).catch(() => {
          reject()
        }).finally(() => { locked = false })
      })
    }
  }
}
// 请求拦截
const beforeRequest = (config: any) => {
  config.retryTimes = 3
  // NOTE  添加自定义头部
  token.value && (config.headers.Authorization = `Bearer ${token.value}`)
  // config.headers['auth_token'] = ''
  return config
}
proxy.interceptors.request.use(beforeRequest)

// 响应拦截器
const responseSuccess = (response: AxiosResponse) => {
  // eslint-disable-next-line yoda
  // 这里没有必要进行判断，axios 内部已经判断
  // const isOk = 200 <= response.status && response.status < 300
  let data = response.data
  if (typeof data === 'object') {
    if (data.code !== 0) {
      return responseFailed({ response } as any)
    }
    data = data.data
  }
  return Promise.resolve(data)
}

const responseFailed = (error: AxiosError) => {
  const { response } = error
  const config = response?.config
  const data = response?.data || {} as any
  if (!window.navigator.onLine) {
    alert('没有网络')
    return Promise.reject(new Error('请检查网络连接'))
  }

  let needRetry = true
  if (response?.status == 404) {
    needRetry = false
  } else if (response?.status == 401) {
    needRetry = false
    // AuthNotFound     = New(40100, "auth not found")
    // AuthExpired      = New(40102, "auth expired")
    if (data.code === 40102 || data.code === 40100) {
      token.value = ''
      return token.update().then(() => {
        return requestRetry(0, response!)
      })
    }
  }
  if (!needRetry) {
    return Promise.reject(data || response)
  };
  return requestRetry(1000, response!)
}

const requestRetry = (delay = 0, response: AxiosResponse) => {
  const config = response?.config
  // @ts-ignore
  const { __retryCount = 0, retryDelay = 1000, retryTimes } = config;
  // 在请求对象上设置重试次数
  // @ts-ignore
  config.__retryCount = __retryCount + 1;
  // 判断是否超过了重试次数
  if (__retryCount >= retryTimes) {
    return Promise.reject(response?.data || response?.headers.error)
  }
  if (delay <= 0) {
    return proxy.request(config as any)
  }
  // 延时处理
  return new Promise<void>((resolve) => {
    setTimeout(() => {
      resolve();
    }, delay)
  }).then(() => {
    return proxy.request(config as any)
  })
}


proxy.interceptors.response.use(responseSuccess, responseFailed)

interface data {
  json?: any
  query?: any
  form?: any
  header?: any
}

function transData(d: data) {
  let opts = { params: d.query, data: {}, headers: {} as any }
  if (d.form) {
    opts.data = d.form
    opts.headers['content-type'] = 'application/x-www-form-urlencoded'
  }
  if (d.json) {
    opts.data = d.json
    opts.headers['content-type'] = 'application/json'
  }
  if (d.header) {
    opts.headers = Object.assign(opts.headers, d.header)
  }
  return opts
}

export const webapi = {
  Get<T>(url: string, req: data): Promise<T> {
    return proxy.request<T, any>(Object.assign({ method: 'get', url: url }, transData(req)))
  },
  Head<T>(url: string, req: data): Promise<T> {
    return proxy.request<T, any>(Object.assign({ method: 'head', url: url }, transData(req)))
  },
  Delete<T>(url: string, req: data): Promise<T> {
    return proxy.request<T, any>(Object.assign({ method: 'delete', url: url }, transData(req)))
  },

  Post<T>(url: string, req: data): Promise<T> {
    return proxy.request<T, any>(Object.assign({ method: 'post', url: url }, transData(req)))
  },
  Put<T>(url: string, req: data): Promise<T> {
    return proxy.request<T, any>(Object.assign({ method: 'put', url: url }, transData(req)))
  },
  Patch<T>(url: string, req: data): Promise<T> {
    return proxy.request<T, any>(Object.assign({ method: 'patch', url: url }, transData(req)))
  },
}

export default webapi
