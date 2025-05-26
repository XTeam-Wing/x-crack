package brute

import (
	"context"
	"fmt"
	"time"
)

// BruteItem 表示一个爆破任务项
type BruteItem struct {
	AllowBlankUsername bool              `json:"allow_blank_username"` // 是否允许空用户名
	AllowBlankPassword bool              `json:"allow_blank_password"` // 是否允许空密码
	Type               string            `json:"type"`                 // 服务类型 (ssh, ftp, mysql, etc.)
	Target             string            `json:"target"`               // 目标地址
	Username           string            `json:"username"`             // 用户名
	Password           string            `json:"password"`             // 密码
	Port               int               `json:"port"`                 // 端口
	Context            context.Context   `json:"-"`                    // 上下文
	Timeout            time.Duration     `json:"timeout"`              // 超时时间
	Extra              map[string]string `json:"extra"`                // 额外参数
}

// BruteResult 表示爆破结果
type BruteResult struct {
	Item           *BruteItem             `json:"item"`
	Success        bool                   `json:"success"`
	Error          error                  `json:"error,omitempty"`
	ResponseTime   time.Duration          `json:"response_time"`
	Banner         string                 `json:"banner,omitempty"`
	Finished       bool                   `json:"finished"`        // 是否完成
	UserEliminated bool                   `json:"user_eliminated"` // 用户是否被排除
	ExtraInfo      map[string]interface{} `json:"extra_info,omitempty"`
}

// String 返回结果的字符串表示
func (r *BruteResult) String() string {
	status := "FAIL"
	if r.Success {
		status = "SUCCESS"
	}
	return fmt.Sprintf("[%s] %s://%s:%s@%s:%d", status, r.Item.Type, r.Item.Username, r.Item.Password, r.Item.Target, r.Item.Port)
}

// BruteCallback 爆破回调函数类型
type BruteCallback func(item *BruteItem) *BruteResult

// ResultCallback 结果回调函数类型
type ResultCallback func(result *BruteResult)

// Config 爆破配置
type Config struct {
	// 并发控制
	TargetConcurrent int `json:"target_concurrent"` // 目标并发数
	TaskConcurrent   int `json:"task_concurrent"`   // 任务并发数

	// 延迟控制
	MinDelay time.Duration `json:"min_delay"` // 最小延迟
	MaxDelay time.Duration `json:"max_delay"` // 最大延迟

	// 超时设置
	Timeout time.Duration `json:"timeout"` // 连接超时

	// 重试设置
	MaxRetries int `json:"max_retries"` // 最大重试次数

	// 停止条件
	OkToStop           bool `json:"ok_to_stop"`          // 成功后是否停止
	FinishingThreshold int  `json:"finishing_threshold"` // 完成阈值

	// 字典设置
	UserDict     []string `json:"user_dict"`      // 用户字典
	PassDict     []string `json:"pass_dict"`      // 密码字典
	UserDictFile string   `json:"user_dict_file"` // 用户字典文件
	PassDictFile string   `json:"pass_dict_file"` // 密码字典文件

	// 其他设置
	SkipEmptyPassword  bool          `json:"skip_empty_password"`  // 跳过空密码
	SkipEmptyUsername  bool          `json:"skip_empty_username"`  // 跳过空用户名
	AllowBlankUsername bool          `json:"allow_blank_username"` // 允许空用户名
	AllowBlankPassword bool          `json:"allow_blank_password"` // 允许空密码
	OnlyNeedPassword   bool          `json:"only_need_password"`   // 只需要密码
	CustomCallback     BruteCallback `json:"-"`                    // 自定义回调

	// 扫描范围
	PortRange    string `json:"port_range"`    // 端口范围
	ExcludePorts []int  `json:"exclude_ports"` // 排除端口
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		TargetConcurrent:   50,
		TaskConcurrent:     10,
		MinDelay:           time.Millisecond * 100,
		MaxDelay:           time.Millisecond * 500,
		Timeout:            time.Second * 10,
		MaxRetries:         3,
		OkToStop:           false,
		FinishingThreshold: 10,
		SkipEmptyPassword:  true,
		SkipEmptyUsername:  true,
		OnlyNeedPassword:   false,
		PortRange:          "",
		ExcludePorts:       []int{},
	}
}
