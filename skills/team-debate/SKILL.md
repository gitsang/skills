---
name: team-debate
description: Use when coordinating team-mode or multi-agent debate for design decisions, first-principle admission, architecture tradeoffs, conceptual boundaries, or any request that says core concepts must be debated with team members. Enforces one-issue-at-a-time debate, two-round minimums, explicit verdicts, timely debate-record updates, and derived-boundary handling.
---

# Team Debate

把团队辩论当成**准入流程**，不是把多个代理意见拼成摘要。

核心原则：**一次只辩论一个议题；每个议题至少两轮；每轮至少两个成员；结论及时落文档；未准入的内容必须明确降级、移除或改写。**

## When to Use

Use this skill when any of these apply:

- 用户要求“与团队成员充分辩论”“多视角评审”“核心概念必须辩论”。
- 需要决定某条原则、架构边界、模块职责或设计命题是否准入。
- team mode 可用，且需要 `team_*` 成员分别从不同视角输出观点。
- 需要把辩论结论写入 `docs/design/*_debate.md` 或类似决策记录。
- 需要区分“第一性原理 / 强制派生边界 / non-goal / 实现细节”。

Do **not** use this for:

- 单点实现、格式修复、机械重命名。
- 用户明确只要快速答案，不要过程。
- 已经有明确决策，只需要执行。

## What This Skill Prevents

常见失败模式：

1. **打包辩论**：把 P1-P8 或多个架构问题一次性交给团队，导致每个结论都没有充分准入。
2. **摘要代替裁决**：只汇总观点，不给 accept / rewrite / reject / downgrade。
3. **降级被误读为可选**：把 mandatory derived boundary 写成“实现建议”。
4. **成员输出缺轮次**：只有一轮，或第二轮没有回应第一轮分歧。
5. **文档滞后**：辩论已闭合，但 `_debate.md` 和主设计文档没有同步更新。
6. **实现泄漏进原则**：把 API、字段、Go 类型、存储、重试参数、UI 文案写成第一性原理。
7. **不可用成员阻塞全局**：某个成员 errored 或未 claim，导致议题无法推进。

## Debate Contract

每个议题必须有清晰合同：

```text
Issue: 单一候选命题，不能夹带其他候选。
Round 1: 至少两个成员独立判断。
Round 2: 至少两个成员基于 Round 1 重新判断。
Verdict: accept / rewrite / reject / downgrade 之一。
Placement: 写入第一性原理、派生边界、non-goal、场景检验或移除。
Wording: 给出最终可落文档措辞。
Exclusions: 明确哪些实现细节不得进入该层级。
```

如果用户指定更高标准（例如每轮 3-4 个成员），以用户标准为准。

## Team Setup

### 1. Assign roles, not generic reviewers

每个成员必须代表不同压力方向，例如：

- `runtime` / `gateway-runtime`：执行路径、性能、自治、故障降级。
- `control-plane` / `manager-control-plane`：管理面状态、操作反馈、审计、用户视图。
- `security` / `security-oidc`：凭据、OIDC、token、权限、攻击面。
- `critic` / `first-principles-critic`：概念层级、是否泄漏实现、是否应降级。

不要让所有成员都回答同一个泛泛问题“怎么看”。给他们明确视角。

### 2. Use task descriptions with required output shape

每个 team task 都应包含：

```text
Round: 第几轮。
Single issue only: 候选命题。
Known context: 已准入 / 已降级的相关命题。
Forbidden scope: 不要讨论哪些相邻议题。
Return format: position, bottom-level proposition, reasoning,
implementation leaks, verdict, suggested wording/placement.
```

### 3. Do not over-block on one member

如果用户要求是“每轮至少两个成员”，两个独立成员完成即可推进。

当成员 errored、长时间 pending、或 task 被删除但消息已到：

- 已收到的有效观点可以纳入记录。
- 不要让一个不可用成员阻塞整个议题。
- 如果当前轮不足两个观点，创建替代任务给可用成员。
- 如果删除了已输出成员的 task，告知其输出已收到，不必补 completion。

## Workflow

### 1. State the current issue

在每轮开始前写清楚：

```text
当前只辩论 Pn：<候选命题>。
本轮不讨论 Pn+1 或其他候选，只允许作为边界备注出现。
```

### 2. Dispatch Round 1

Round 1 要求成员独立回答：

- 这是不是底层命题？
- 是否依赖已接受的原则？
- 是否组件 / 拓扑 / 实现形状过重？
- 应 accept / rewrite / reject / downgrade？
- 如果降级，是否仍然 mandatory？放在哪里？

### 3. Synthesize before Round 2

Round 2 不是重复投票。必须把 Round 1 的主要分歧或共识反馈给成员：

```text
Round 1 consensus: A and B recommend downgrade; C recommends rewrite.
Re-evaluate whether any reason remains to keep it in first-principles.
Return final wording, placement, and risk if downgraded.
```

### 4. Close the issue immediately

一旦一个议题达到轮次和成员数要求：

1. 形成最终 verdict。
2. 立刻更新 debate record。
3. 立刻同步主设计文档。
4. 再启动下一个候选。

不要等所有候选辩论完再集中写文档；这样最容易丢结论或混入旧结论。

### 5. Preserve mandatory derived requirements

“不进入第一性原理”不等于“不重要”。

如果团队结论是 downgrade but mandatory，主文档必须用明确措辞：

```text
<要求> 是由 Pn / Pm 推导出的强制派生边界，不是独立第一性原理。
```

并写入合适章节：

- Derived boundary
- Management/control-plane guardrail
- Security/privacy non-goal
- Runtime consistency invariant
- Scenario or implementation constraint

## Debate Record Template

```markdown
## 议题：<议题名称>

### 第 1 轮辩论

#### 观点1：<member>

- **立场**：...
- **底层命题**：...
- **理由**：...
- **建议位置**：...
- **边界**：...

### 第 2 轮辩论

#### 观点K：<member>

- **最终判断**：accept / rewrite / reject / downgrade。
- **建议措辞**：“...”。
- **风险**：...

### 结论

> 最终措辞。

该结论必须 / 不得 ...；具体 <实现细节> 均属于后续实现边界。
```

## Decision Heuristics

### Accept into first principles

Only accept if the statement remains true after replacing:

- Redis / DB / MQ / WebSocket / push / pull
- API / route / RPC / UI
- Go type / schema / field name
- single gateway / multi gateway topology
- product workflow wording

It should define stable domain semantics, not implementation behavior.

### Rewrite before accepting

Rewrite when the idea is primitive but wording leaks implementation.

Examples:

- “Gateway owns sessions” → “请求路径执行权与管理控制意图必须分离”。
- “复用 `Authentication`” → “认证事实必须遵守统一语义契约”。

### Downgrade to mandatory derived boundary

Downgrade when the requirement is essential but derives from accepted principles.

Typical derived boundaries:

- control action authorization / idempotency / auditability
- distributed convergence / revocation semantics
- least-credential / least-disclosure management-plane boundary
- UI state truthfulness for pending / partial / failed operations

Always label mandatory boundaries as mandatory.

### Reject or remove

Reject when the statement is:

- merely an implementation choice
- a UI workflow
- a field list
- a product scenario
- redundant with an accepted principle and not useful as a derived guardrail

## Verification Checklist

Before reporting completion:

- [ ] Each issue has exactly one candidate focus.
- [ ] Each issue has at least two rounds.
- [ ] Each round has at least two member outputs or the user-approved higher threshold.
- [ ] Each issue has an explicit verdict.
- [ ] Accepted principles appear in the target principle chapter.
- [ ] Downgraded mandatory requirements appear in derived-boundary / non-goal
      sections with “mandatory” language.
- [ ] Debate record and main document are synchronized.
- [ ] No stale bundled-language remains: `P1-P8`, `all principles`,
      `final set`, `准入 N 条`.
- [ ] No rejected / downgraded candidate remains as a peer principle.
- [ ] Implementation details are excluded from principle wording.

Use text search for stale phrases after editing. If markdown docs are changed,
run formatter and markdown lint when available.

## Red Flags

Stop and fix if you notice:

- “先把所有原则都让大家评一下” — split into one issue per principle.
- “大家都差不多同意” — still need verdict, wording, placement, exclusions.
- “降级到实现建议” — if mandatory, call it mandatory derived boundary.
- “这只是历史记录，不用改” — stale records can contradict final docs.
- “成员还没完成但已有两个输出” — okay to proceed only if user threshold is
  met.
- “第 2 轮直接复制第 1 轮问题” — round 2 must respond to round 1.

## Team Tool Notes

If team mode is enabled and `team_*` tools are available, use them directly.

Useful pattern:

1. `team_task_create` for each member/round.
2. `team_send_message` to ask members to claim tasks.
3. `team_task_list` / `team_status` to monitor progress.
4. Record peer messages as they arrive.
5. Delete or replace stale pending tasks only when enough valid outputs exist or
   a substitute task has been created.

Do not inspect config files just to prove team mode exists; the presence of
`team_*` tools is enough.
