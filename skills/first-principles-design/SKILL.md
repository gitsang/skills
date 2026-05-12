---
name: first-principles-design
description: Use when designing or revising software architecture, middleware boundaries, policy models, product concepts, or any feature whose requirements are expanding and need a stable conceptual model before implementation. Especially useful when the user asks whether to merge or split modules, worries about duplicated responsibility, says “第一性原理”, “概念一致性”, “边界”, “会不会变成脚本系统”, or wants a design document under docs/design.
---

# First-Principles Design

把设计工作从“这个功能怎么做”提升到“这类需求到底属于什么概念”。

核心原则：**通过持续追问和多视角检验找到稳定原语，再决定模块边界；先解释为什么，再写 YAML、接口或代码。**

这不是一次性产出设计结论的 skill。它要求 agent 主动和用户、subagent、team 成员来回交流，持续扩充场景、反例和边界条件，直到概念模型能解释已知需求，也能约束未来需求。

## When to Use

Use this skill when the task has any of these signals:

- 用户在讨论模块应该合并还是拆分。
- 同一个能力似乎可以放在多个地方配置。
- 需求正在从一个具体功能扩散成一组规则、动作、状态或策略。
- 用户担心系统会变成“大脚本”“万能规则引擎”“筐”。
- 需要把讨论整理成 `docs/design/` 下的设计文档。
- 需要长期指导原则，而不是只要一个当前可行方案。

Do **not** use this for:

- 已经边界清楚的单点实现任务。
- 纯代码修 bug。
- 只有 UI 文案或样式调整的任务。
- 用户明确只要快速答案，不要设计过程。

## What This Skill Prevents

通用代理在架构设计里常见失误：

1. 看到两个功能有重复字段，就马上合并成一个大模块。
2. 看到概念不同，就完全拆开，导致底层能力重复实现。
3. 把 action 写进模块名，比如 `failBan`，后来发现 action 不止 ban。
4. 直接给配置草案，却没说明触发原因、执行阶段和状态模型。
5. 追着新需求改字段，最后做出一个隐形脚本系统。

这个 skill 的作用是强制先建立概念坐标系。

## Workflow

### 1. Gather context before judging

不要凭直觉判断“该合并”或“该拆分”。先读现有系统：

- 现有模块和中间件的职责。
- 配置结构和命名习惯。
- 哪些模块已经分离了“事实生产”和“执行拦截”。
- 是否有历史上被删除或失败的大模块。
- 现有文档里是否已经定义了相关概念。

如果问题复杂，至少并行做两类检索：

- 代码库检索：找当前实现边界、配置样例、旧实现教训。
- 外部模型检索：找类似系统的概念划分，例如 OPA、Envoy、Kubernetes admission、WAF、fail2ban。

如果当前环境支持 subagent 或 team，不要只靠自己想：

- 派一个代码库观察者找现有边界和历史包袱。
- 派一个外部参考观察者找相似系统的概念模型。
- 派一个反方观察者专门攻击你的合并/拆分方案。
- 如果 team mode 可用，让不同成员分别扮演“产品使用者”“实现者”“运维者”“未来需求提出者”。

每个代理都要被要求返回：场景、反例、命名风险、用户可能困惑的地方。不要只让他们总结文件。

### 2. Expand scenarios through questions

在提出方案前，至少问一轮场景问题。好的问题不是“你要 A 还是 B”，而是暴露边界：

- “如果 action 以后不止 ban，这个名字还成立吗？”
- “同一个能力如果能在两个地方配置，用户会怎么判断放哪里？”
- “这个规则依赖的是请求前事实，还是响应后事实？”
- “有没有一个反例会迫使我们把两个模块重新合并？”
- “有没有一个未来动作会让当前命名失效？”

问题要一轮一轮问。每轮只问最能改变设计方向的问题。用户回答后，更新概念模型，不要假装原方案一直正确。

如果用户表达了担心，例如“会不会变成脚本系统”，把它当作设计约束，不是闲聊。

### 3. Extract first-principle primitives

把需求拆成少量稳定原语。常用问题：

```text
phase      什么时候发生？请求前、响应后，还是未来请求前？
subject    观察谁？IP、用户、Token、租户、路径、资源？
condition  因为什么触发？权限、流量、失败、异常、配额？
state      依赖什么历史状态？计数器、桶、封禁标记、会话、授权结果？
action     做什么？allow、deny、ban、throttle、challenge、notify、mark？
```

如果这五个问题答不清楚，不要继续设计字段。

### 4. Separate trigger reason from action

模块边界按触发原因划分，不按 action 划分。

例子：

- `accessControl` 触发原因是“权限/准入”。
- `trafficControl` 触发原因是“流量过高”。
- `failureControl` 触发原因是“失败模式异常”。

同一个 action 可以出现在多个模块里：

- `deny` 可以因为无权限、超限、封禁状态触发。
- `notify` 可以因为失败异常、流量异常、权限异常触发。
- `throttle` 可以是基础流量策略，也可以是失败后的临时惩罚。

这不是职责重叠。职责由 trigger reason 决定。

### 5. Stress-test with divergent cases

拿用户给出的需求和代理收集到的场景做压力测试：

```text
如果 action 增加一种，新模型是否仍然成立？
如果 condition 增加一种，应该归到哪个控制类别？
如果某个能力可在两个地方配置，是不是说明边界错了？
如果删除一个模块，另一个模块是否会被迫承担不属于它的职责？
```

必须主动寻找会破坏当前设计的例子。找不到时，也要说明“目前未发现破坏性反例”，不要说“这个设计很完美”。

### 6. Decide what unifies and what stays typed

优先选择：

```text
Shared Core + Typed Surfaces
```

也就是：底层统一原语，用户侧保持清晰类别。

不要轻易选择：

```text
One Big Generic Policy Script
```

除非用户真的需要完整策略语言，并愿意承担调试、顺序、副作用和安全成本。

也不要选择：

```text
Many Unrelated Middlewares
```

否则 subject 选择、matcher、counter、state、action 会到处重复，用户也会不知道能力该配在哪里。

### 7. Name by domain, not by first action

命名要能容纳未来合理扩展。

Good:

- `trafficControl`：可以包含 rate limit、quota、concurrency、delay。
- `failureControl`：可以包含 ban、throttle、challenge、notify。
- `accessControl`：表达准入和权限边界。

Bad:

- `failBan`：把 ban 写死，但 failure 的 action 不止 ban。
- `gatewayProtection`：太宽，容易变成筐。
- `riskControl`：范围过大，容易吸收所有风控需求。

命名要拿未来场景测试：

- 如果 action 从 `ban` 变成 `throttle`，名字是否仍然成立？
- 如果观察对象从失败变成所有响应状态，名字是否过窄？
- 如果用户看到这个名字，能否知道配置该放这里还是放别处？

### 8. Model phases explicitly

很多设计混乱来自 phase 没说清楚。

常见 phase：

- `preRequest`：请求前判断，适合 access control、baseline traffic control。
- `postResponse`：响应后观察，适合 failure control、audit。
- `enforceState`：根据历史状态拦截未来请求，比如 ban、temporary throttle。

如果一个模块同时涉及观察和执行，要明确两个阶段。例如 failure control：

```text
postResponse: 观察 401/403/5xx，更新失败计数
preRequest: 检查 ban/throttle/challenge 状态并执行
```

### 9. Define ownership rules

用一组判断规则防止用户困惑：

```text
这个请求有没有权限？
  -> accessControl

这个主体流量是否过高？
  -> trafficControl

这个主体失败行为是否异常？
  -> failureControl

只是记录或通知？
  -> action 或 audit，不单独作为判定模块
```

文档里必须写清楚：同一个 action 可以被多个控制类别使用，但 condition 的归属不能重复。

### 10. Keep the first version bounded

设计可以给出长期概念，但第一版实现必须克制。

示例：

```text
第一版做：
  trafficControl.rateLimit
  failureControl.observeStatusCode
  failureControl.ban
  failureControl.throttle

第一版不做：
  通用脚本表达式
  任意 action 编排
  通知系统
  challenge 页面
  WAF 规则语言
  旧模块大重写
```

## Output Structure

当用户要求整理设计文档时，默认写到：

```text
docs/design/<topic>.md
```

推荐结构：

```markdown
# <Concept A> / <Concept B> / <Concept C> 设计

## 背景
说明为什么现有讨论会发散，以及这份文档要解决什么概念问题。

## 第一性原则
列出稳定原语，例如 phase / subject / condition / state / action。

## 总体模型
说明采用 Shared Core + Typed Surfaces，或者其他模型。

## <Concept A>
回答它解决什么问题、输入是什么、action 是什么、不应该负责什么。

## <Concept B>
同上。

## <Concept C>
同上。

## 易混问题
直接回答用户最担心的问题，例如“failure 的 action 能不能是 rate limit”。

## 为什么不完全合并
说明大一统脚本系统的成本。

## 为什么不完全拆散
说明重复实现和用户困惑的成本。

## 职责归属规则
给出遇到新需求时的判断表。

## 第一版边界
明确做什么、不做什么。

## 命名
说明为什么采用当前命名，拒绝哪些命名。
```

## Quality Bar

完成前检查：

- 是否有一个能解释未来需求的稳定原语模型？
- 是否至少通过一轮用户追问扩充了场景或确认了边界？
- 是否使用 subagent/team 或等价的多视角分析检查过反例？如果没有，是否说明问题足够简单？
- 是否区分了 trigger reason 和 action？
- 是否明确了 phase？
- 是否解释了为什么不完全合并、也不完全拆散？
- 是否给出了用户以后能用的职责归属规则？
- 是否避免了“万能”“强大”“灵活”这类空话？
- 是否有第一版边界，避免变成脚本系统？

## Interaction Rules

这个 skill 的关键不是“写出漂亮文档”，而是逼近正确概念。执行时必须保持交互：

1. **先复述设计张力。** 例如：“这里的冲突不是 rateLimit 能不能作为 action，而是 baseline traffic policy 和 adaptive failure policy 是否混淆。”
2. **每轮只问一个高价值问题。** 问能改变模块边界的问题，不问偏好题。
3. **让用户的担心进入模型。** 用户说“怕变成脚本系统”，文档里就要有“不做脚本系统”的边界。
4. **向 subagent/team 要反例。** 不要只要赞同意见。明确要求他们攻击当前命名、拆分和 phase 模型。
5. **综合后再下结论。** 先列出发现，再给推荐，不要把推荐伪装成事实。
6. **记录被放弃的方案。** 写清楚为什么不完全合并，为什么不完全拆散。

### Useful Subagent / Team Prompts

代码库观察者：

```text
CONTEXT: We are designing first-principles boundaries for <topic>.
GOAL: Find existing module boundaries, naming patterns, config shapes, and historical complexity traps.
REQUEST: Return concrete files, current concepts, overlaps, and constraints. Also list one naming or boundary risk.
```

反方观察者：

```text
CONTEXT: Current proposal is <summary>.
GOAL: Break the proposal before we document it.
REQUEST: Find scenarios where the proposed split confuses users, duplicates configuration, or collapses into a script system. Return counterexamples and a better boundary if you see one.
```

未来需求观察者：

```text
CONTEXT: We need a concept model that survives future feature requests.
GOAL: Generate plausible next-year requirements.
REQUEST: List 5 future requirements and test whether each fits the current primitives: phase, subject, condition, state, action.
```

## Example Pattern

用户问：

> failure 的 action 是否可以有 rateLimit？这和 rateLimit 是否冲突？是否应该和 access control 合并？

好的回答不是直接说“可以”或“不可以”，而是先拆：

```text
trafficControl.rateLimit 是 baseline policy。
failureControl.throttle 是失败触发后的 adaptive policy。
accessControl 是权限准入。
三者共享 policy core，但用户侧保持 typed controls。
```

然后再给配置命名建议：

```yaml
failureControl:
  rules:
    - observe:
        statusCode: "401"
      subject: ip
      threshold: 5
      action:
        throttle:
          limit: 10
          window: 1m
          duration: 10m
```

这里用 `throttle`，不用 `rateLimit`，因为这是临时惩罚状态，不是基础流量规则。
