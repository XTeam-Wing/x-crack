# x-crack

一个功能强大且灵活的弱口令扫描工具，使用Go语言开发，支持多种协议的爆破攻击。

## 特性

- 🚀 **高性能**: 支持高并发爆破，可自定义并发数
- 🔧 **多协议支持**: SSH, FTP, MySQL, PostgreSQL, Redis, MongoDB, HTTP等
- 📝 **灵活配置**: 支持配置文件、命令行参数等多种配置方式
- 🎯 **精确控制**: 支持延迟控制、重试机制、成功停止等
- 📊 **多种输出**: 支持文本、JSON、CSV等输出格式
- 🔌 **可扩展**: 支持自定义协议处理器和回调函数
- 📦 **模块化**: 可作为库导入到其他Go项目中
- 📊 **详细输出**: 支持JSON格式输出和进度显示
- 🔌 **SDK调用**: 可作为底层引擎被其他平台调用

## 快速开始

### CLI 使用

```bash
# 基本用法
x-crack -t 192.168.1.100:22 --service ssh

# 批量扫描
x-crack -t targets.txt --service ssh,ftp,mysql

# 自定义字典
x-crack -t 192.168.1.100:3306 --service mysql -u admin,root -p admin,123456

# 高级配置
x-crack -t 192.168.1.0/24:22 --service ssh --threads 50 --timeout 10s --delay 1s

# 使用配置文件
x-crack --config config.yaml
```

### 配置文件

x-crack 支持YAML格式的配置文件，可以通过 `--config` 参数指定：

```yaml
# config.yaml
version: "1.0.0"
debug: false
log_level: "info"

# 爆破设置
brute:
  target_concurrent: 50    # 目标并发数
  task_concurrent: 1       # 任务并发数
  timeout: "10s"          # 连接超时
  max_retries: 3          # 最大重试次数
  
  # 默认字典
  default_user_dict:
    - "admin"
    - "root"
    - "administrator"
  
  default_pass_dict:
    - "123456"
    - "password"
    - "admin"

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

### SDK 使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/XTeam-Wing/x-crack/pkg/brute"
)

func main() {
    engine := brute.NewEngine()
    
    config := &brute.Config{
        Targets:     []string{"192.168.1.100:22"},
        Services:    []string{"ssh"},
        Threads:     10,
        Timeout:     time.Second * 10,
        Usernames:   []string{"root", "admin"},
        Passwords:   []string{"123456", "admin"},
    }
    
    results := make(chan *brute.Result, 100)
    
    ctx := context.Background()
    err := engine.Start(ctx, config, results)
    if err != nil {
        panic(err)
    }
    
    for result := range results {
        if result.Success {
            fmt.Printf("成功: %s %s:%s\n", result.Target, result.Username, result.Password)
        }
    }
}
```

## 项目结构

```
x-crack/
├── cmd/                    # CLI应用入口
├── internal/              # 核心业务逻辑
│   ├── brute/            # 暴力破解引擎
│   ├── protocols/        # 协议实现
│   ├── dictionaries/     # 字典管理
│   └── config/          # 配置管理
├── pkg/                  # 对外SDK接口
├── configs/             # 配置文件
├── test/               # 测试文件
└── README.md
```

## 许可证

GPL-3.0 License
