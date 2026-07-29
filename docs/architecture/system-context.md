# ShipGuard System Context

## 1. 核心参与者

### Developer

开发者提交代码、查看流水线结果并申请发布。

### Approver

审批 staging 或 production 环境的高风险发布。

### Platform Engineer

维护服务目录、环境配置、发布策略和平台组件。

### SRE

维护 SLI、SLO 和错误预算，处理事故并完成复盘。

## 2. ShipGuard 核心进程

### ShipGuard API

ShipGuard API 负责：

- 服务目录管理；
- 创建和查询发布；
- 接收审批结果；
- 查询审计记录；
- 查询事故和复盘。

### ShipGuard Controller

ShipGuard Controller 负责：

- 推进发布状态机；
- 创建 GitOps 变更；
- 查询 Argo CD 和 Argo Rollouts 状态；
- 执行 SLO 检查；
- 触发晋级、终止或回滚；
- 收集故障证据。

## 3. 交付与部署组件

### GitHub Actions

GitHub Actions 负责：

- 执行单元测试；
- 构建容器镜像；
- 执行 Trivy 安全扫描；
- 生成 SBOM；
- 将制品信息提交给 ShipGuard。

### GitOps Repository

GitOps Repository 保存 Kubernetes 的期望部署状态。

### Argo CD

Argo CD 监控 GitOps Repository，并将期望状态同步到 Kubernetes。

### Argo Rollouts

Argo Rollouts 执行 Canary 或 Blue-Green 渐进式发布。

## 4. 可观测性与故障证据组件

### Prometheus

Prometheus 保存应用、发布和基础设施指标，并为发布门禁和 SLO 检查提供数据。

### Loki

Loki 保存应用日志、平台日志和发布相关日志。

### Tempo

Tempo 保存分布式链路追踪数据，用于定位慢请求和下游调用异常。

### Cluster Inspector

Cluster Inspector 分析 Kubernetes 资源状态、Warning Event 和常见故障，输出诊断证据。

### AI Summary Layer

AI Summary Layer 基于已经收集的指标、日志、Trace、事件和 Runbook 生成摘要与解释。

AI 不负责审批发布、修改 GitOps 配置、执行回滚或直接操作生产集群。
