# ShipGuard

**ShipGuard｜云原生发布与稳定性工程平台**

ShipGuard 是一个面向 SRE、DevOps、平台工程和云原生研发场景的发布控制与稳定性平台。

## 核心链路

代码提交
→ 单元测试
→ 构建镜像
→ 安全扫描
→ GitOps 部署
→ 灰度发布
→ SLO 检查
→ 自动回滚
→ 变更记录与复盘

## 核心原则

- 使用 Go 实现发布控制平面。
- 使用 PostgreSQL 保存发布状态、审批和审计数据。
- GitOps 仓库是 Kubernetes 部署期望状态的唯一来源。
- Argo CD 负责同步部署。
- Argo Rollouts 负责金丝雀和蓝绿发布。
- Prometheus、Grafana、Loki、Tempo 提供可观测性。
- AI 只负责总结、检索和解释，不执行生产操作。

## 当前阶段

Stage 0：项目立项、环境检查和架构初始化。
