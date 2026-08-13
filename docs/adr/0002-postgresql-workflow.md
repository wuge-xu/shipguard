# ADR-0002：使用 PostgreSQL 保存工作流和审计数据

- Status: Accepted
- Date: 2026-08-13

## Context

ShipGuard 的发布、审批、审计、SLO 检查和事故数据具有强关联关系。

一次发布状态变化通常需要同时更新 ReleaseRun，并追加 ReleaseEvent 或 Approval 记录。

如果这些数据分别写入多个非事务存储，就可能出现状态已经变化但审计事件没有保存的问题。

## Decision

ShipGuard 使用 PostgreSQL 作为核心工作流和审计数据库。

PostgreSQL 保存：

- 服务目录；
- 环境；
- 发布运行；
- 发布状态转换事件；
- 审批记录；
- 安全扫描结果；
- SLO 评估结果；
- 事故和 Postmortem；
- Runbook 元数据。

## Worker 领取任务

第一版 Controller Worker 使用 PostgreSQL 行锁领取待处理任务。

核心机制采用 SELECT ... FOR UPDATE SKIP LOCKED。

这样多个 Worker 可以并发扫描任务，同时跳过已经被其他 Worker 锁定的记录，避免重复领取。

## Alternatives

### Redis 作为核心队列和状态存储

Redis 适合高吞吐队列，但 ShipGuard 的核心问题不是单纯消息吞吐，而是发布状态、审批和审计之间的一致性。

如果同时使用 Redis 和 PostgreSQL 保存核心状态，会额外引入双写一致性问题。

因此第一版不使用 Redis 作为发布工作流的核心事实来源。

### 独立消息系统

Kafka、NATS 或 RabbitMQ 可以提供更强的异步能力，但在当前阶段会增加运维和一致性复杂度。

当任务吞吐或跨服务事件规模明显增长时，再评估引入独立消息系统。

## Transactional Outbox

后续需要对外发送事件时，ShipGuard 将使用 Transactional Outbox。

业务状态和 OutboxEvent 在同一个 PostgreSQL 事务中提交，再由独立 Worker 异步投递外部事件。

这样可以避免数据库状态已经提交，但外部事件发送失败导致的信息丢失。

## Consequences

优点：

- 发布状态和审计记录可以事务提交；
- 可以使用唯一约束保证幂等；
- 支持复杂历史查询和事故审计；
- 可以使用行锁实现多 Worker 并发领取；
- 后续可以自然扩展 Transactional Outbox。

代价：

- 数据库轮询会增加一定查询压力；
- 必须正确设计事务范围和锁粒度；
- 高吞吐场景下可能需要独立消息系统。
