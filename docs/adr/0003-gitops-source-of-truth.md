# ADR-0003：GitOps 仓库作为部署期望状态唯一事实来源

- Status: Accepted
- Date: 2026-08-13

## Context

ShipGuard 需要同时处理数据库中的发布记录、Git 中的部署配置和 Kubernetes 中的实际运行状态。

如果数据库和 GitOps Repository 都可以独立定义期望部署版本，就会形成多个事实来源。

一旦多个来源发生不一致，平台将无法可靠判断应该恢复哪一个状态。

## Decision

Kubernetes 部署期望状态只由 GitOps Repository 定义。

ShipGuard PostgreSQL 不保存可直接应用到 Kubernetes 的完整期望配置。

数据库只记录：

- Release ID；
- 源代码 Commit SHA；
- 镜像 Digest；
- GitOps Commit SHA；
- 发布状态；
- 审批记录；
- 回滚原因；
- 审计事件。

ShipGuard 通过修改 GitOps Repository 发起持久化部署变更。

Argo CD 负责将 Git 中声明的期望状态同步到 Kubernetes。

## 发布异常时的两层恢复机制

ShipGuard 将发布故障处理分为快速止损和持久状态恢复两层。

### 第一层：快速 Abort

当 Canary 发布期间 Prometheus 分析结果失败时，ShipGuard 或 Argo Rollouts 立即停止继续晋级。

流量保持或恢复到 Stable ReplicaSet，避免故障版本继续扩大影响范围。

这一层的目标是快速止损，不负责改变 GitOps Repository 中的最终期望状态。

### 第二层：Git Revert

当确认需要回滚时，ShipGuard 创建正式回滚操作，并修改 GitOps Repository，使期望版本恢复到已知稳定版本。

新的 Git Commit 会记录回滚版本、原因和关联 Release ID。

Argo CD 随后将该 Git 状态重新同步到 Kubernetes。

因此，集群内的 Abort 用于快速恢复流量，Git Revert 用于恢复持久事实来源。

## Alternatives

### 仅直接修改 Kubernetes 资源

这种方式恢复速度快，但会绕过 GitOps Repository，导致集群实际状态与 Git 中期望状态发生漂移。

### 仅依赖 Git Revert

这种方式保持 GitOps 一致性，但等待 Git 提交和 Argo CD 同步可能增加故障影响时间。

## Consequences

优点：

- 故障版本可以快速停止扩散；
- GitOps Repository 最终仍保持唯一事实来源；
- 每一次回滚都有 Git Commit 和 ShipGuard 审计记录；
- 可以区分临时流量止损和正式版本恢复。

代价：

- ShipGuard 必须协调 Rollout 状态和 GitOps 状态；
- 需要处理 Abort 成功但 Git Revert 失败的部分失败场景；
- 回滚操作必须保证幂等，避免重复生成 Git 提交。
