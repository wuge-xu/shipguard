# ADR-0004：AI 仅作为只读总结与解释层

- Status: Accepted
- Date: 2026-08-13

## Context

ShipGuard 涉及生产发布、审批、回滚和 Kubernetes 操作。

如果直接让大模型修改 GitOps 配置、执行回滚或操作集群，会引入不可预测、难审计和权限失控风险。

## Decision

AI 只处理 ShipGuard 已经收集好的确定性证据。

允许输入的证据包括：

- ReleaseEvent；
- Prometheus 查询结果；
- Loki 日志；
- Tempo Trace；
- Kubernetes Event；
- Argo Rollouts 状态；
- Cluster Inspector 报告；
- Runbook。

AI 可以：

- 生成故障摘要；
- 整理事故时间线；
- 关联不同来源的证据；
- 检索相关 Runbook；
- 生成 Postmortem 草稿；
- 标记事实、推断、未知信息和建议。

AI 不可以：

- 审批发布；
- 修改 GitOps Repository；
- 执行回滚；
- 删除或重启工作负载；
- 修改 SLO；
- 获取长期生产写权限。

## Consequences

优点：

- 发布和回滚逻辑保持确定性；
- AI 输出可以被人工验证；
- 降低提示注入和错误操作风险；
- AI 失败不会直接改变生产环境。

代价：

- AI 无法独立完成事故处置；
- 必须先建设高质量证据收集链路；
- AI 输出必须明确区分事实和推断。
