# AgentSocial 调研报告

> 项目代号：AgentSocial
> 日期：2026-02-12（v0.2 更新）

---

## 一、核心概念验证

### 1.1 你的想法的数学本质

假设有 N 个人，每两人之间互相有意向的概率为 p（如 1%）：

- **传统模式（你主动筛选）**：你需要评估 N 次 → 成本 O(N)
- **反转模式（对方先筛选）**：对方先评估你 → 只有 N×p 个人通过 → 你只需评估 N×p 次 → 成本 O(N×p)
- **双 Agent 模式**：双方 Agent 各自独立评估 → 只有双方都通过的才进入真人环节 → 真人只需处理 N×p² 次 → 成本 O(N×p²)

**结论：Agent 双向预筛选将 10000 人的筛选量从 10000 次降低到 ~1 次（在 p=1% 时），效率提升 10000 倍。**

### 1.2 市场验证 —— 已有先行者

| 项目 | 类型 | 规模 | 与我们的差异 |
|------|------|------|------------|
| **MoltMatch** | AI Agent 相亲 | 活跃 | 仅限约会场景，Agent 只做破冰，不做深度匹配 |
| **Volar** | AI 化身约会 | 已上线 | Avatar 对话→人类接管，但不是基于个人 Agent |
| **Tinder AI Twin** | AI 双胞胎测试 | 测试中 | 大平台内部功能，不可扩展 |
| **Tezi (Max)** | AI 自动招聘 | $9M 融资 | 仅招聘方视角，不是双向 Agent 匹配 |
| **MoltBook** | AI Agent 社交网络 | 260万 Agent | 公共论坛模式，非点对点匹配 |

**关键市场缺口：没有一个平台同时实现了——**
1. 每个人都有自己的 AI Agent 替身
2. Agent-to-Agent 的自主匹配和筛选
3. 多场景通用（招聘 + 社交 + 合作）
4. 从 Agent 匹配到真人建联的完整闭环

### 1.3 技术可行性验证

基于对 claw 服务器上 OpenClaw v2026.2.9 的实机调研：

| 能力 | 可行性 | 说明 |
|------|--------|------|
| SOUL.md 定义 Agent 人格 | ✅ 已验证 | 已有完善的人格系统 |
| Skill 系统扩展能力 | ✅ 已验证 | SKILL.md 格式成熟，支持 API 调用 |
| Cron 定时任务 | ✅ 已验证 | 用户可自行配置，支持分钟级定时扫描 |
| 多 Agent 隔离 | ✅ 已验证 | 每个 Agent 独立 workspace/sessions/memory |
| Agent-to-Agent 通信 | ⚠️ 实验性 | agentToAgent 工具存在但默认关闭 |
| Embedding 向量匹配 | ✅ 可用 | 平台侧使用 SQLite + sqlite-vec |
| Telegram 通信 | ✅ 已验证 | 已在 claw 上成功配对 |

---

## 二、竞品深度分析

### 2.1 MoltBook 生态

MoltBook 是目前最大的 AI Agent 社交网络（260万+ Agent），但它本质是一个**公共论坛**（类似 Reddit），不是点对点匹配平台。

**其通信架构值得借鉴的点：**
- **Heartbeat 系统**：Agent 每 4 小时自动轮询平台获取指令
- **AgentCard**：Agent 的身份名片，包含能力、技能、联系方式
- **MoltSpeak 协议**：Ed25519 签名 + 加密的结构化消息，比自然语言节省 54% 的 token
- **HOL Registry**：跨注册中心的 Agent 发现，72000+ Agent，14 个注册中心

**其局限性：**
- 公共论坛模式，不适合隐私敏感的招聘/社交匹配
- 严重的 prompt injection 安全问题
- 404 Media 报道了未加密数据库漏洞
- 被 MIT Technology Review 称为"peak AI theater"

### 2.2 Agent-to-Agent 协议现状

| 协议 | 定位 | 成熟度 | 适用性 |
|------|------|--------|--------|
| **Google A2A** | Agent 间协作 | 标准化中（Linux 基金会） | 适合跨系统 Agent 发现 |
| **Anthropic MCP** | Agent 接工具 | 成熟（97M+ 月下载） | 适合平台 API 接入 |
| **Agent Relay** | 直接消息传递 | 原型阶段 | 简单的消息收发 |
| **MoltSpeak** | 加密结构化消息 | 原型阶段 | 安全通信参考 |

### 2.3 Embedding 匹配技术现状

- **LinkedIn EBR**：双塔神经网络 + 余弦相似度 + IVFPQ 近似最近邻搜索
- **Resume2Vec**：Transformer(BERT/RoBERTa) 生成简历/JD 向量，比传统 ATS 高 15.85%
- **Iris Dating**：面部吸引力向量（Deep Metric Learning），用于约会匹配
- **我们选用**：OpenAI `text-embedding-3-large`（裁剪到 256 维度），基于关键词（非句子），在平台侧计算

---

## 三、方案评估：方案一 vs 方案二

### 方案一：在用户自己的 OpenClaw 上添加社交 Skill

**实现方式：** 用户安装一个 `agentsocial` skill，配置一个 `SOCIAL.md`，即可让自己的 OpenClaw Agent 参与匹配网络。

| 维度 | 评分 | 说明 |
|------|------|------|
| 用户接入成本 | ⭐⭐⭐⭐⭐ | 一条命令安装 skill，编辑一个 md 文件 |
| 上下文丰富度 | ⭐⭐⭐⭐⭐ | Agent 已有用户的完整记忆/人格/偏好 |
| 隐私控制 | ⭐⭐⭐⭐ | 数据留在用户本地，只发送必要信息 |
| AI Native 感 | ⭐⭐⭐⭐⭐ | 编辑 md → 安装 skill → Agent 自动工作 |
| 开发成本 | ⭐⭐⭐⭐⭐ | 只需开发 skill + 平台 API |
| 专注度 | ⭐⭐⭐ | 通过 session 隔离 + 独立 cron job 解决 |
| 生态兼容性 | ⭐⭐⭐⭐⭐ | 兼容 OpenClaw 全部已有能力 |
| 推广传播性 | ⭐⭐⭐⭐ | "一个 skill 让你的 AI 替你社交" |

### 方案二：Fork OpenClaw 做独立的社交 Agent 项目 → ❌ 不采用

理由：用户体验差（需安装新系统）、缺失上下文（新 Agent 不了解用户）、维护 fork 成本高、无法借势 OpenClaw 生态。

### 结论：采用方案一

---

## 四、角色模型设计决策

### 问题

原设计中"主动方"（持有资源、等待匹配的一方）和"被动方"（主动扫描寻找的一方）命名有严重歧义——"被动方"反而是最活跃的。

### 解决方案：Beacon / Radar 模式

| 模式 | 含义 | 典型场景 |
|------|------|---------|
| **Beacon（灯塔模式）** 🔦 | 发布需求，等待被发现和匹配 | 招聘方发布 JD、热门社交对象、项目发起人 |
| **Radar（雷达模式）** 📡 | 主动扫描平台，寻找匹配目标 | 求职者、追求方、寻找合伙人 |

**设计要点：**
- 同一用户可以在不同任务中使用不同模式
- 模式选择由用户通过自然语言与 Agent 沟通确定
- Agent 应主动与用户讨论清楚每个任务的模式选择
- 招聘场景中通常招聘方是 Beacon、求职者是 Radar，但也有例外（如主动挖人）

---

## 五、关键参考资料

- [OpenClaw Official Docs](https://docs.openclaw.ai/)
- [OpenClaw GitHub](https://github.com/openclaw/openclaw) (145K+ stars)
- [MoltBook](https://moltbook.com/) (260万+ Agent)
- [MoltMatch](https://moltmatch.xyz/) (Agent 相亲)
- [Google A2A Protocol](https://a2a-protocol.org/)
- [Anthropic MCP](https://modelcontextprotocol.io/)
- [Volar Dating](https://www.volar.dating/) (AI Avatar 约会)
- [Tezi AI Recruiting](https://tezi.ai/) (自主 AI 招聘)
- [Agent Relay Protocol](https://agent-relay.onrender.com/)
- [MoltSpeak Protocol](https://www.moltspeak.xyz/)
- [HOL Agent Registry](https://hol.org/registry/moltbook)
- [LinkedIn EBR](https://www.linkedin.com/blog/engineering/platform-platformization/using-embeddings-to-up-its-match-game-for-job-seekers)
- [arXiv: AI Agents with DIDs](https://arxiv.org/html/2511.02841v1)
