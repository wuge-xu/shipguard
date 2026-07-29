# ShipGuard Project Charter

## 1. 项目使命

ShipGuard 用于控制云原生应用从代码提交到生产发布的完整生命周期。

平台需要保证发布过程可控制、可观察、可回滚、可审计，并在异常发生后保存完整故障证据。

## 2. 核心问题

ShipGuard 主要解决以下问题：

1. 谁正在发布哪个服务和版本？
2. 版本是否通过测试和安全扫描？
3. 哪些环境需要人工审批？
4. GitOps 配置是否已经同步到 Kubernetes？
5. 灰度版本的错误率和延迟是否正常？
6. 服务是否有足够错误预算继续发布？
7. 发布失败后能否及时停止和回滚？
8. 发布、审批和回滚是否完整可追踪？
9. 故障后能否关联指标、日志、链路和集群事件？

## 3. 项目范围

首个完整版本必须覆盖：

- 服务目录；
- 发布状态机；
- GitHub Actions；
- 镜像构建；
- Trivy 安全扫描；
- GitOps 多环境部署；
- Argo CD；
- Argo Rollouts；
- 发布审批；
- SLO 发布门禁；
- 自动终止和回滚；
- 发布审计；
- Prometheus、Grafana、Loki、Tempo；
- 错误预算；
- Chaos Mesh；
- Cluster Inspector 接入；
- Postmortem；
- Runbook 检索；
- AI 故障证据总结。

## 4. 非目标

ShipGuard 第一版不负责：

- 自行实现 Kubernetes；
- 自行实现镜像仓库；
- 替代 Argo CD；
- 替代 Argo Rollouts；
- 让 AI 直接执行生产操作；
- 一开始开发复杂前端；
- 一开始支持大量云平台和多集群。

## 5. 核心演示场景

故障版本进入金丝雀发布后导致错误率升高。

Prometheus 检测异常，Argo Rollouts 自动停止发布，ShipGuard 回退 GitOps 配置并保存完整审计记录。

平台随后收集指标、日志、Trace、Kubernetes 事件和 Cluster Inspector 报告，生成 Postmortem 草稿。
