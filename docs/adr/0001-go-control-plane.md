# ADR-0001：使用 Go 构建发布控制平面

- Status: Accepted
- Date: 2026-08-13

## Context

ShipGuard 需要长期协调发布状态、审批、GitOps、Argo CD、Argo Rollouts、Prometheus 和 Kubernetes。

如果只使用 GitHub Actions 串联这些工具，发布流程会分散在 CI YAML 中，难以统一保存状态、实现恢复逻辑和提供完整审计。

## Decision

ShipGuard 使用 Go 实现自己的发布控制平面。

控制平面拆分为两个主要进程：

- shipguard-api：负责同步 API 请求、查询、发布申请和审批操作。
- shipguard-controller：负责异步推进发布状态机、协调外部系统和处理失败恢复。

GitHub Actions、Argo CD 和 Argo Rollouts 仍负责各自擅长的执行能力，但发布流程的业务状态由 ShipGuard 控制。

## Alternatives

### 仅使用 GitHub Actions

优点是实现简单，但状态分散在工作流中，不适合复杂审批、重试、恢复和审计。

### 自己实现完整 GitOps 和灰度控制器

会重复实现 Argo CD 和 Argo Rollouts 已经解决的问题，增加大量无价值复杂度。

## Consequences

优点：

- 可以统一实现发布状态机；
- 可以实现事务、幂等、重试和失败恢复；
- 可以集中保存审批和审计记录；
- 可以作为 Go 平台后端独立扩展和测试。

代价：

- 需要自行设计清晰的领域模型；
- 需要处理并发、状态一致性和外部系统失败；
- 必须为核心状态转换建立完整自动化测试。
