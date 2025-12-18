# example-go-hoststat
HostStat 是一个轻量级的主机状态采集工具，用于获取服务器的基础信息、网络状态以及磁盘与 IO 性能指标，适用于主机监控、运维巡检和基础设施可观测性场景。

## 核心功能

- **系统指标采集**：实时采集 CPU、内存、磁盘、网络和 I/O 等关键系统指标
- **安全 API 访问**：通过 token 验证机制确保 API 接口的安全访问
- **轻量级设计**：低资源占用，适合在各种服务器环境中部署
- **嵌入式资源**：静态资源（HTML、favicon.ico）嵌入到可执行文件中，简化部署
- **模板渲染**：使用 Go 的 html/template 进行页面渲染

## 技术栈

- **Go 语言**：高性能后端开发语言
- **标准库**：使用 Go 标准库实现 HTTP 服务、模板渲染、JSON 处理等功能
- **嵌入式资源**：使用 embed 包将静态资源嵌入到可执行文件中
- **安全机制**：基于 HTTP-only Cookie 和请求头验证的安全机制

## 安装和运行

### 编译

```bash
go build -o hoststat-go main.go
```

### 运行

```bash
./hoststat-go
```

服务将在 `:8080` 端口启动。

## API 接口

### 基础信息接口

- **URL**: `/base`
- **Method**: `GET`
- **Description**: 获取系统基础信息
- **Response**: JSON 格式的系统基础信息

### 当前状态接口

- **URL**: `/current`
- **Method**: `GET`
- **Description**: 获取系统当前运行状态
- **Response**: JSON 格式的系统当前状态信息

### CPU 使用率接口

- **URL**: `/top/cpu/ps`
- **Method**: `GET`
- **Description**: 获取 CPU 使用率最高的进程
- **Response**: JSON 格式的进程列表

### 内存使用率接口

- **URL**: `/top/mem/ps`
- **Method**: `GET`
- **Description**: 获取内存使用率最高的进程
- **Response**: JSON 格式的进程列表

## 安全机制

### Token 生成和验证

- **生成**: 当访问主页时，服务会生成包含请求头信息的 token 并设置为 HTTP-only Cookie
- **验证**: API 接口通过验证请求中的 token 来确保访问安全
- **机制**: token 包含请求头的 JSON 序列化数据，并使用 base64 编码

### 安全中间件

所有 API 接口都通过 `SecureMiddleware` 进行保护，确保只有携带有效 token 的请求才能访问。

## 开发指南

### 添加新的 API 接口

1. 在 `handles/hander.go` 中添加新的处理函数
2. 在 `main.go` 中注册新的路由，并应用 `SecureMiddleware`

### 扩展系统指标采集

在 `handles/hander.go` 中实现新的指标采集函数，并在相应的处理函数中调用。

## 配置说明

当前版本主要通过代码中的常量和变量进行配置，包括：

- **服务端口**: `:8080`
- **Token 有效期**: 24 小时
- **Cookie 配置**: HTTP-only、SameSite=Strict、Secure=false