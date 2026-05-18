---
name: team-debate
description: Use when coordinating team-mode or multi-agent debate for design decisions, first-principle admission, architecture tradeoffs, conceptual boundaries, or any request that says core concepts must be debated with team members. Enforces one-issue-at-a-time debate, role-separated rounds, lead verdicts over majority voting, timely records, and derived-boundary handling.
---

# Team Debate

核心原则：**一次只辩论一个议题；先独立判断，再受控复判；多数票只是信号，lead 才能裁决；结论必须小步落档。**

## 何时使用

当以下任一情况成立时，使用这个技能：

- **具有实际权衡取舍的战略决策**: 诸如“自建还是购买”、技术栈选择或定价策略等决策，实际上存在着不同视角之间的张力。
  如果由产品、工程和业务人员分别扮演不同的角色，就能发现那些单一角色无法弥合的真正分歧。
- **需要对抗性审查的内容**: 如果你要撰写一篇技术博客文章、一份政策文件或一份提案，而这些内容将面临真正的审查，
  那么通过辩论循环（一个代理人寻找逻辑缺陷，另一个代理人检查缺失的上下文）来运行它，比单次生成产生更好的结果。
- **研究综合分析可能存在偏差**: 在总结有关有争议话题的研究时，让一个代理人负责寻找支持某种立场的证据，
  让另一个代理人负责寻找反对该立场的证据，比要求一个代理人“做到公平”更能产生平衡的结果。
- **大规模代码审查**: 三个角色各异的智能体 “安全审查员”、“性能审查员” 和 “可维护性审查员”，
  分别可以发现代码库中不同类型的问题。这可以作为人工审查的补充，而非替代。
- **情景规划** 将代理人视为乐观、悲观和中立的情景规划者，可以为业务规划或风险评估生成更丰富的结果集。

不要用于：

- **简单的查询任务**：如果答案很明确，争论只会增加成本，而不会提高质量。“法国的首都是哪里？”这个问题不需要三个人参与。
- **纯粹的生成性任务**：写俳句或生成模板不需要辩论。因为不存在需要达成共识的真理。
- **高频、低风险查询**：如果您每天运行数千个低风险查询，那么多代理辩论的成本和延迟可能就无法证明其合理性。
- **在对速度要求极高的任务中**：辩论会增加延迟。实时应用程序通常无法承受这种延迟。

实用筛选：当单个输出的质量具有实际后果，问题具有真正的权衡取舍，并且额外成本因利害关系而合理时，可以使用多主体辩论。

## 先决条件

需要拥有 subagent 或 team/swarm 等多 Agent 调用能力/工具。

当 team 可用时，优先使用 team 执行。

## 开始执行

请按照以下步骤设计辩论

### 步骤 1：定义问题类型，配置团队成员

不同的问题需要不同的辩论形式。请具体说明你想要改进/讨论的地方：

| 问题类型                   | 推荐配置                               |
| -------------------------- | -------------------------------------- |
| 事实准确性（研究问题）     | 多个具有独立推理能力的个体进行辩论     |
| 创意质量（文案撰写、设计） | 具有不同审美价值观的多个角色           |
| 风险评估                   | 乐观主义者 + 怀疑主义者 + 中立的综合者 |
| 技术决策                   | 与决策相关的领域专家角色               |
| 伦理问题                   | 具有不同利益相关者视角的代理人         |

对于大多数应用场景而言，3-5 个智能体是比较实用的范围。
两个智能体可以进行辩论，但如果没有打破僵局的机制，往往会陷入僵局。
超过五个智能体则容易产生嘈杂的输出，因为智能体会把时间浪费在总结其他智能体的观点上，而不是贡献新的分析。
三个智能体（两个观点相反的智能体和一个综合分析智能体）是最常见且有效的配置。

每个成员都必须代表不同压力方向，
不要让所有成员都回答同一个泛泛问题“怎么看”，要给他们明确视角。
如果有可能，可以选择使用不同的模型进行辩论。

### 步骤二：编写不同的系统提示

每个代理的系统提示需要：

- 树立鲜明的个人形象：不仅仅是一个名字，而是一套价值观、优先事项和行为准则。
- 给出明确指示：这位代理人在辩论中试图达成什么目标？
- 定下基调：这位经纪人是怀疑论者？做事有条理者？还是富有创造力者？
- 具体说明要质疑的内容：该代理人应该对哪些类型的索赔提出异议？

以下是一个产品决策辩论的具体例子：

```
Agent 1 — The Product Optimist

You are a product strategist who strongly believes in user value and growth.
You tend to favor shipping features quickly to learn from real users.
When evaluating proposals, you focus on user benefit, market opportunity,
and speed to feedback. You're skeptical of over-engineering and analysis paralysis.
Challenge arguments that prioritize internal concerns over user outcomes.
```

```
Agent 2 — The Engineering Skeptic

You are a senior engineer with extensive experience in technical debt and system failures.
You believe most product decisions are made without enough regard for maintainability,
scalability, and long-term cost. When evaluating proposals, identify the technical risks,
implementation complexity, and things likely to break at scale.
Push back on vague technical claims and unrealistic timelines.
```

```
Agent 3 — The Synthesizer

You have read the arguments from both the product optimist and the engineering skeptic.
Your job is not to pick a side but to find the most defensible position that acknowledges
the strongest points from each perspective. Produce a concrete recommendation with
explicit trade-offs. Be specific — avoid vague compromises.
```

### 步骤 3：选择循环结构

你需要决定进行几轮辩论以及辩论顺序。

> 大多数实际应用都采用一次性辩论和一次修改的方式。这种方式能够以极低的成本实现大部分质量改进。

#### 一次性辩论（最经济）：

- 所有 Agent 对最初的问题均独立作答。
- 每个 Agent 阅读其他 Agent 的回复，并提出一次修改意见。
- Synthesizer 生成最终答案。

#### 多轮辩论（最全面）：

- 所有 Agent 独立做出反应。
- Agent 阅读回复并反驳——重点关注分歧。
- Agent 对反驳作出回应。
- 重复 N 轮。
- Synthesizer或人工评估最终结论。

#### 异步辩论（对异步系统而言最现实）：

- Agent 1 做出回应。
- Agent 2 回应 Agent 1。
- Agent 3 回应 Agent 1, Agent 2。
- 继续循环直至收敛或达到整数极限。

### 步骤 4：定义停止条件

你需要制定一个何时停止的规则。常见选项：

- 固定轮数：无论是否收敛，在 N 轮后停止。
- 共识阈值：当所有参与者对同一答案达成一致时停止。
- 人工触发：运行多轮，直到人工认为输出足够好为止。
- 发散标志：如果智能体陷入循环（重复争论而没有新信息），则提前停止。

对于大多数生产应用场景，固定轮数（1-3）加上最后一步的合成是最实用的方法。

## 状态持久化

每个 Agent 都应该将自己的观点历史，记录到文件。

默认情况下，应该按照以下格式记录：

1. 所有记录存放到 `.team-debate/` 目录。
2. 每个议题存放到一个单独的目录 `.team-debate/<议题名称>`。
3. 每个 Agent 的记录存储在 `.team-debate/<议题名称>/<agent-name>.json` 文件中。
4. 每一轮辩论和最终的结论存储到 `.team-debate/<议题名称>/DEBATE.md`

## 增强模式

### 非对称信息模式

给不同的参与者提供不同的原始文件或数据，然后让他们展开讨论。这种方法适用于竞争分析、尽职调查或情景规划等任务。

### 收敛性检查模式

在每一轮讨论后添加一个简单的检查，看看代理人之间是否仍然存在实质性分歧。

使用廉价模型或快速模型，进行收敛性检查，既能降低成本，又能避免不必要的辩论轮次。

## 常见失败

### 智能体锚定在看到的第一个回复上

症状：如果智能体 2 在形成自己的观点之前就阅读了智能体 1 的回答，那么辩论就会陷入僵局，最终达成一致。

修复方案：
在执行第一次回复时，应该尽可能只提供有限的信息，并且在表达上减少主观判断，如果可以，应该禁止 Agent 在第一次回复时使用工具探索目录。

### 代理人过于相似

症状：Agent 之间很快达成一致。“辩论” 实际上只是一轮 “说得对，我同意”。

修复方案：
重写系统提示，赋予代理人真正不同的职责和价值观。
添加明确的指令，例如 “默认情况下，你对其他代理人的推理持怀疑态度。只有在证据确凿的情况下才必须同意。”
或者添加一个专门的 “反方辩护人” Agent，其唯一职责就是找出漏洞。

### 没有收敛机制

症状：Agent 们争论不休。五轮争论下来，他们还在重复同样的话。

改进方案：
在每轮结束后增加收敛性检查。
引入一个协调者代理，它可以宣布僵局，并要求代理们提出 “在分歧存在的情况下他们的最佳立场”。
或者，增加一个固定的轮数限制，并强制进行综合讨论。

### Synthesis 不能打破僵局

症状：综合分析得出含糊不清的结论：“双方都有道理。这取决于你的优先事项。”

改进方案：
明确授权合成器即使在不确定情况下也要提出建议。
在其系统提示中添加：“您必须提出具体的建议。‘视情况而定’是不可接受的。请说明您的建议以及建议会改变的条件。”

### 将共识视为正确

症状：你相信辩论结果，因为“三位参与者都达成了一致”。但他们达成一致的是错误的。

修正：
辩论中达成的共识比单一主体输出更可靠，但并非绝对可靠。
对于高风险输出，务必添加特定领域的验证。
对于事实性主张，添加事实核查步骤，查询外部来源。
对于技术性建议，添加一个在辩论结束后运行的独立验证主体。
