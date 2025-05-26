# x-crack

一个功能强大且灵活的弱口令扫描工具，使用Go语言开发，支持多种协议的爆破攻击。

## 🌟 特性

- 🚀 **高性能**: 支持高并发爆破，可自定义并发数和超时控制
- 🔧 **多协议支持**: SSH, FTP, Telnet, MySQL, PostgreSQL, Redis, MongoDB, HTTP/HTTPS, SMB, VNC, SNMP, IMAP, POP3, SMTP 等16种协议
- 📝 **灵活配置**: 支持YAML配置文件、命令行参数等多种配置方式
- 🎯 **精确控制**: 支持延迟控制、重试机制、成功停止等高级功能
- 📊 **多种输出**: 支持文本、JSON、CSV等输出格式
- 🔌 **可扩展**: 模块化设计，支持自定义协议处理器
- 📦 **SDK支持**: 可作为库导入到其他Go项目中
- ⏱️ **超时控制**: 每个协议独立的连接超时设置
- 🔄 **容错机制**: 支持失败重试和断点续传

## 📋 支持的协议

| 协议 | 默认端口 | 状态 | 描述 |
|------|----------|------|------|
| SSH | 22 | ✅ | Secure Shell 协议 |
| FTP | 21 | ✅ | File Transfer Protocol |
| Telnet | 23 | ✅ | 远程登录协议 |
| MySQL | 3306 | ✅ | MySQL 数据库 |
| PostgreSQL | 5432 | ✅ | PostgreSQL 数据库 |
| Redis | 6379 | ✅ | Redis 内存数据库 |
| MongoDB | 27017 | ✅ | MongoDB 文档数据库 |
| HTTP | 80, 8080, 8000, 8888 | ✅ | HTTP Basic Auth |
| HTTPS | 443, 8443 | ✅ | HTTPS Basic Auth |
| SMB | 445, 139 | ✅ | Server Message Block |
| VNC | 5900-5902 | ✅ | Virtual Network Computing |
| SNMP | 161 | ✅ | Simple Network Management Protocol |
| IMAP | 143, 993 | ✅ | Internet Message Access Protocol |
| POP3 | 110, 995 | ✅ | Post Office Protocol |
| SMTP | 25, 587, 465 | ✅ | Simple Mail Transfer Protocol |
| RDP | 3389 | ⚠️ | Remote Desktop Protocol (待完善) |

## 🚀 快速开始

### 安装

```bash
# 从源码编译
git clone https://github.com/XTeam-Wing/x-crack.git
cd x-crack
go build -o x-crack ./cmd/x-crack

# 或直接下载发布版本
# 下载对应平台的二进制文件到 PATH 目录
```

### 基本使用

```bash
# 单目标单协议爆破
./x-crack -target 192.168.1.100:22 -protocol ssh -username root -password 123456

# 多用户名密码组合
./x-crack -target 192.168.1.100:22 -protocol ssh -usernames root,admin -passwords 123456,password

# 批量目标扫描
./x-crack -targets 192.168.1.100:22,192.168.1.101:3306 -protocols ssh,mysql

# 从文件读取目标和凭据
./x-crack -target-file targets.txt -user-file users.txt -pass-file passwords.txt -protocol ssh

# 高级配置
./x-crack -target 192.168.1.100:22 -protocol ssh \
  -target-concurrent 50 -task-concurrent 30 \
  -timeout 5s -retries 3 -delay 100ms

# 输出到文件
./x-crack -target 192.168.1.100:22 -protocol ssh \
  -output results.json -format json -verbose
```

### 配置文件

x-crack 支持YAML格式的配置文件，通过 `-config` 参数指定：

```yaml
# config.yaml
version: "1.0.0"
debug: false
log_level: "info"

# 爆破设置
brute:
  # 并发控制
  target_concurrent: 50    # 目标并发数
  task_concurrent: 1       # 任务并发数
  
  # 超时设置
  timeout: "10s"          # 连接超时
  
  # 重试设置
  max_retries: 3          # 最大重试次数
  
  # 停止条件
  ok_to_stop: false       # 成功后是否停止
  
  # 默认字典
  default_user_dict:
    - "admin"
    - "root"
    - "administrator"
    - "user"
    - "test"
    - "guest"
  
  default_pass_dict:
    - "123456"
    - "password"
    - "admin"
    - "root"
    - "123456789"
    - "12345678"

# 输出设置
output:
  format: "text"          # 输出格式: text, json, csv
  verbose: false          # 详细输出
  show_failed: false      # 显示失败尝试

# 代理设置 (可选)
proxy:
  enabled: false
  type: "http"           # http, https, socks5
  address: ""
```

使用配置文件：

```bash
# 使用默认配置文件
./x-crack -config config.yaml

# 配置文件 + 命令行参数 (命令行参数优先级更高)
./x-crack -config config.yaml -target 192.168.1.100:22 -protocol ssh
```

## 🔧 命令行参数

```bash
Usage:
  ./x-crack [flags]

目标设置:
   -target string       目标主机 (例如: 192.168.1.1:22)
   -targets string[]    目标主机列表 (逗号分隔)
   -target-file string  包含目标主机的文件
   -port int            目标端口
   -ports string        端口范围 (例如: 22,3389,1433-1434)
   -port-file string    包含端口的文件
   -protocol string     使用的协议 (ssh,mysql,ftp等)
   -protocols string[]  协议列表 (逗号分隔)

认证设置:
   -u, -username string    认证用户名
   -usernames string[]     用户名列表 (逗号分隔)
   -uf, -user-file string  包含用户名的文件
   -p, -password string    认证密码
   -passwords string[]     密码列表 (逗号分隔)
   -pf, -pass-file string  包含密码的文件
   -userpass-file string   包含用户名:密码组合的文件

爆破设置:
   -target-concurrent int  目标并发数 (默认: 50)
   -task-concurrent int    每个目标的任务并发数 (默认: 30)
   -delay string           请求间延迟 (例如: 100ms)
   -timeout string         每个请求的超时时间 (默认: 10s)
   -retries int            失败重试次数 (默认: 3)
   -ok-to-stop             首次成功认证后停止 (默认: true)

输出设置:
   -output string  输出文件路径
   -format string  输出格式 (text,json,csv) (默认: text)
   -v, -verbose    详细输出
   -d, -debug      调试模式
   -silent         静默模式
   -no-color       禁用彩色输出
   -show-failed    显示失败的认证尝试

其他设置:
   -config string  配置文件路径
   -version        显示版本信息
```

## 📚 SDK 使用

x-crack 可以作为 Go 库导入到其他项目中：

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/XTeam-Wing/x-crack/pkg/brute"
    "github.com/XTeam-Wing/x-crack/pkg/config"
)

func main() {
    // 创建配置
    cfg := &config.Config{
        Brute: config.BruteConfig{
            TargetConcurrent: 10,
            TaskConcurrent:   5,
            Timeout:          "10s",
            MaxRetries:       3,
            OkToStop:         true,
        },
    }
    
    // 创建爆破器
    builder := brute.NewBuilder().
        WithTargets([]string{"192.168.1.100:22"}).
        WithProtocols([]string{"ssh"}).
        WithUsernames([]string{"root", "admin"}).
        WithPasswords([]string{"123456", "password"}).
        WithConfig(cfg)
    
    // 设置结果回调
    builder.WithResultCallback(func(result *brute.Result) {
        if result.Success {
            fmt.Printf("成功: %s %s:%s\n", 
                result.Target, result.Username, result.Password)
        } else {
            fmt.Printf("失败: %s %s:%s - %v\n", 
                result.Target, result.Username, result.Password, result.Error)
        }
    })
    
    // 执行爆破
    ctx := context.Background()
    err := builder.BatchBruteWithConfig(ctx)
    if err != nil {
        panic(err)
    }
}
```

## 🏗️ 项目结构

```
x-crack/
├── cmd/
│   └── x-crack/            # CLI 应用入口
│       └── main.go
├── pkg/
│   ├── brute/              # 爆破引擎核心
│   │   ├── engine.go       # 爆破引擎
│   │   ├── builder.go      # 构建器模式
│   │   └── types.go        # 类型定义
│   ├── config/             # 配置管理
│   │   └── config.go       # 配置加载和保存
│   ├── protocols/          # 协议实现
│   │   ├── ssh.go          # SSH 协议
│   │   ├── mysql.go        # MySQL 协议
│   │   ├── ftp.go          # FTP 协议
│   │   ├── ...             # 其他协议
│   │   └── register.go     # 协议注册器
│   └── utils/              # 工具函数
│       └── utils.go
├── dict/                   # 默认字典文件
│   ├── usernames.txt
│   ├── passwords.txt
│   └── combo.txt
├── test/                   # 测试文件
├── config.yaml             # 默认配置文件
├── go.mod
├── go.sum
└── README.md
```

## ⚡ 性能特性

- **高并发**: 支持目标级和任务级并发控制
- **内存优化**: 流式处理，支持大规模扫描
- **超时控制**: 每个协议独立的连接超时
- **重试机制**: 智能重试失败的连接
- **资源管理**: 自动连接池和资源释放
- **限流控制**: 可配置的请求速率限制

## 🔒 安全说明

**免责声明**: 本工具仅用于授权的安全测试和教育目的。请确保：

1. 仅在您拥有或已获得明确授权的系统上使用
2. 遵守当地法律法规和公司政策
3. 不要用于恶意攻击或非法入侵
4. 建议在测试环境中验证功能

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/new-protocol`)
3. 提交更改 (`git commit -am 'Add new protocol support'`)
4. 推送到分支 (`git push origin feature/new-protocol`)
5. 创建 Pull Request

### 添加新协议

要添加新协议支持，请：

1. 在 `pkg/protocols/` 目录下创建新的协议文件
2. 实现 `ProtocolHandler` 接口
3. 在 `register.go` 中注册新协议
4. 添加相应的测试

## 📄 许可证

本项目采用 GPL-3.0 许可证。详见 [LICENSE](LICENSE) 文件。

## 🔗 相关链接

- [GitHub 仓库](https://github.com/XTeam-Wing/x-crack)
- [问题反馈](https://github.com/XTeam-Wing/x-crack/issues)
- [Wiki 文档](https://github.com/XTeam-Wing/x-crack/wiki)

---

⭐ 如果这个项目对您有帮助，请给我们一个星标！
