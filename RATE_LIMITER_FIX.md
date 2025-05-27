# 限速器和并发设置修复说明

## 修复的问题

### 1. 限流器配置错误
**原问题**: 
- 限流器的参数配置不正确，导致限流效果无效
- `rate.NewLimiter` 的参数使用有误

**修复方案**:
- 正确使用 `rate.Every(config.MinDelay)` 作为速率参数
- 使用 `config.TargetConcurrent` 作为突发容量参数
- 添加了动态调整限流器的功能

### 2. 并发控制逻辑混乱
**原问题**:
- `TargetConcurrent` 和 `TaskConcurrent` 的概念混淆
- 缺少全局并发控制，只有局部信号量
- 信号量管理不当，可能导致泄漏

**修复方案**:
- 明确区分全局并发数(`TargetConcurrent`)和单目标并发数(`TaskConcurrent`)
- 添加 `globalSem` 信号量控制全局并发
- 改进信号量的获取和释放逻辑，避免泄漏

### 3. 配置验证不足
**原问题**:
- 配置验证过于简单，缺少边界检查
- 没有合理性检查和警告

**修复方案**:
- 增强配置验证逻辑，包括详细的错误信息
- 添加合理性检查，对过高的并发数给出警告
- 验证延迟时间的合理性

### 4. 缺少监控和调试能力
**原问题**:
- 无法动态调整限流器设置
- 缺少并发状态监控
- 日志信息不够详细

**修复方案**:
- 添加 `UpdateRateLimit` 方法支持动态调整
- 添加 `GetRateLimitStatus` 和 `GetConcurrencyStatus` 方法
- 改进日志记录，包含详细的配置信息

## 核心改进

### 1. 并发控制架构
```
全局并发控制 (globalSem) 
    ↓
目标级并发控制 (process.semaphore)
    ↓
限流器控制 (limiter.Wait)
    ↓
实际执行任务
```

### 2. 配置参数说明
- `TargetConcurrent`: 全局最大并发任务数
- `TaskConcurrent`: 单个目标的最大并发任务数  
- `MinDelay`: 每个请求之间的最小延迟
- `MaxDelay`: 最大延迟（用于未来的随机延迟功能）

### 3. 默认配置优化
- 降低默认并发数，避免过度并发导致的问题
- 增加默认延迟时间，减少对目标系统的压力
- 更加保守和安全的默认设置

## 测试验证

创建了完整的测试用例验证：
1. **并发控制测试**: 验证并发数不会超过配置限制
2. **限流器测试**: 验证请求间隔符合延迟设置
3. **动态调整测试**: 验证运行时配置更新功能
4. **状态监控测试**: 验证状态获取功能

## 使用示例

```go
config := &brute.Config{
    TargetConcurrent: 10,                    // 全局最大10个并发
    TaskConcurrent:   3,                     // 每个目标最大3个并发
    MinDelay:         time.Millisecond * 200, // 200ms延迟
    Timeout:          time.Second * 10,       // 10s超时
}

engine, _ := brute.NewEngine(ctx, config)

// 动态调整限流器
engine.UpdateRateLimit(time.Millisecond*500, 20)

// 监控状态
globalUsed, globalTotal, targetUsed, targetTotal := engine.GetConcurrencyStatus()
limit, burst := engine.GetRateLimitStatus()
```

## 性能优化建议

1. **合理设置并发数**: 
   - 全局并发数建议不超过50
   - 单目标并发数建议不超过10

2. **适当的延迟设置**:
   - 对于敏感目标，建议延迟>=500ms
   - 对于内网测试，可以降低到100ms

3. **监控和调整**:
   - 使用状态监控功能观察实际并发情况
   - 根据目标响应情况动态调整参数

## 注意事项

1. 修复后的版本对配置验证更加严格
2. 默认配置更加保守，可能需要根据实际需求调整
3. 新增的监控功能可以帮助优化配置参数
4. 建议在使用前运行测试确认功能正常
