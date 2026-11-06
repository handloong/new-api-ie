# 更新日志

本文件记录 new-api-ie 二开版本的重要变更，按版本倒序排列。

版本号规则：自 `v1.0.2-ie` 起使用 `v主.次.修订-ie` 形式，后缀 `-ie` 用于与上游同号标签区分；`v1.0.1` 及此前的 `v1.0.0-rc.45` ~ `rc.50` 不带后缀，为历史版本，标签保留。二开起点是 `rc.45`。

每次从上游 `QuantumNous/new-api` 合并后都会递增修订号并打 tag，因此本文件中同步类版本可能只有上游改动、不含二开改动。

---

## v1.0.13-ie — 2026-11-06

同步上游 4 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

> 共 4 条，以下全部列出（按提交时间倒序）。

- `2de51a19e` fix(cache): 修复用户和令牌缓存的 Redis 跨槽错误 (#8218)
- `72f3dc7da` fix(pricing): review model price changes before saving
- `e3e74205f` fix(pricing): guide users through model pricing issues
- `3374f251e` feat: forward image complexity and markup for gpt-image-2.6 (#8208)

修正用户与令牌缓存的 Redis 跨槽（CROSSSLOT）错误，避免哈希键落在不同槽位导致读写失败；
定价编辑改为先审阅差异再保存，新增 `price-change-diff.tsx`、`model-pricing-issue-alerts.tsx`
与价格变更快照对比，模型倍率可视化编辑器同步重构；gpt-image-2.6 透传图片复杂度与加价字段
（`relaykit/dto/openai_image.go`）。七套前端语言包均已补齐新增文案。

---

## v1.0.12-ie — 2026-11-05

同步上游 1 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

> 共 1 条，以下全部列出。

- `706144769` fix(auth): 校验 Redis 鉴权缓存完整性 (#8200)

修复 Redis 鉴权缓存中不可解析字段导致的完整性缺陷：token 缓存哈希里存在无法解析的字段、
或混入多余的 `DeletedAt` 字段时，原逻辑每次请求都回退到数据库且永不回写。现在修复路径单独读取
已存签名、丢弃 `DeletedAt`，并把 token 与 user 缓存中非整数的配额计数器重置（合法计数仍保留）；
新增 `model/auth_cache_integrity.go` 与对应测试（含 303 行单元测试、`middleware/auth_test.go` 81 行）。

---

## v1.0.11-ie — 2026-11-04

同步上游 5 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

> 共 5 条，以下全部列出（按提交时间倒序）。

- `ba5f73677` fix(gemini): bill embeddings from upstream usageMetadata (#8182)
- `94b521664` feat(vertex): add per-channel v1beta1 endpoint switch (#8179)
- `83b18fe6c` feat(channel): return created IDs with input mappings (#8164)
- `52c173038` feat(gemini): add Argon fields and reasoning mappings (#8177)
- `9b473a3fe` fix(twofa): restart the failure count after a lockout expires (#8172)

其中 Gemini 嵌入改按上游 `usageMetadata` 计费；Vertex 渠道新增 v1beta1 端点开关（渠道设置项）；
渠道创建接口返回按输入映射生成的 ID（新增 `controller/channel_create_test.go`）；
Gemini DTO 补 Argon 字段与推理映射；两步验证在锁定过期后重置失败计数。

---

## v1.0.10-ie — 2026-11-03

同步上游 3 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

> 共 3 条，以下全部列出（按提交时间倒序）。

- `45746eee4` feat(relay): add opt-in HTTP/2 health checks (#8059)
- `c046559bb` feat(plugins): bill gemini omni by the tokens vertex reports (#8153)
- `651d5a213` fix(xai): preserve token details in stream usage conversion (#6420)

其中 HTTP/2 健康检查新增 `common/constants.go` 开关、`common/init.go` 初始化、`service/http_client.go`
传输层可选启用（默认关闭，沿用原 HTTP/1.1 行为），并新增 140 行传输层测试；vertex-ai 任务插件改为
按 Vertex 上报的 token 明细计费 Gemini Omni；xAI 流式用量转换补回 token details。

---

## v1.0.9-ie — 2026-11-02

同步上游 24 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

> 共 24 条，以下仅列前 10 条（按提交时间倒序）。

- `70e6af642` perf: read and edit request JSON and base64 with moego (#8148)
- `d0abbf6e0` feat(plugins): generate video with Gemini Omni on Vertex AI channels (#8142)
- `ea343c4a8` fix(plugins): charge relayed streams even when the reserve or task insert fails (#8121)
- `fe57cba42` fix(plugins): prefer Kling's returned price (list_price) for Kling 4.0 cash billing (#8125)
- `1424ffdd4` fix(plugins): load plugins that declare their own /{key} address (#8119)
- `8b253de42` fix(plugins): keep Kling 4.0 durations in range and map every video size (#8118)
- `c59f6fd55` feat(plugins): serve task plugins at the vendors' own paths and answer as the vendors do (#8117)
- `37d23dcc8` fix(profile): honor global sidebar visibility (#7429)
- `77436271f` fix(deps): upgrade DOMPurify to 3.4.17 (#8116)
- `883eeb87d` feat(plugins): warn about task models a plugin cannot charge

其余 14 条为 Web/Electron 依赖升级（@codemirror/language、marked、@tanstack/react-query-devtools、
knip、tailwind-merge、fast-uri、electron 41→42、undici、go-jose 4.1.6）以及 Turnstile 密钥保存、
日志轮转后 gin access/recovery 仍写当前文件、登出时卸载已认证布局等修复。

本次共 24 条。

---

## v1.0.8-ie — 2026-10-30

同步上游 1 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

- `a2defd2a8` Merge commit from fork — 敏感信息脱敏时改用 `strings.TrimRight` 一次性裁掉尾部标点（含 `)`），不再逐字符回溯、也不再每遇到一个 `)` 就重算整个 URL 的括号数；被裁掉的标点仍拼回脱敏结果之后

本次共 1 条。

---

## v1.0.7-ie — 2026-10-29

同步上游 8 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

- `e449cc700` feat: sample immediate task successes in model performance metrics (#8067)
- `fcc241cdf` fix: start affinity fallback at the highest channel priority (#8078)
- `b08ee86fe` fix(cloudflare): preserve embedding response data (#8041)
- `26c488c44` fix(relay): allow Azure channels to opt in to Responses WebSocket (#8047) (#8075)
- `55515dbae` fix(relay): terminate streams on exact bare done markers (#8074)
- `49271d2d8` fix(aws): preserve safeguards in Bedrock Invoke requests (#8071)
- `54788117d` fix: count completed non-stream relays after context cancellation (#7844)
- `ad5611282` fix(relay): apply GPT-6 chat parameter rules to gpt-6.1-sol and gpt-6.1-terra (#8064)

本次共 8 条。

---

## v1.0.6-ie — 2026-10-28

同步上游 1 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

- `01da3a55f` perf(relay): stop recompiling param override regexes on every request (#8061)

本次共 1 条。

---

## v1.0.5-ie — 2026-10-26

同步上游 14 个提交，无二开改动。合并时 `README.md` 出现冲突，按约定以二开中文 README 为准保留，上游文档内容不并入；编译与回归测试通过。

**同步上游**

- `34003e9c2` test: stabilize channel affinity cache keys (#8009)
- `0165fef98` fix(relaykit): preserve Gemini text alongside tool calls (#7998)
- `4d962e464` fix(relaykit): preserve tool argument numbers in Claude responses (#7996)
- `f1569d1ea` fix(pricing): classify dynamic per-request billing correctly (#8010)
- `bc795cc11` fix:修复stop为非字符串和[]string的协议转换问题 (#7935)
- `9823ebc7e` fix(billing): bill cache_creation_input_tokens reported in prompt_tokens_details (#7909)
- `e835b1bfc` chore(deps): bump axios from 1.18.1 to 1.20.0 in /web (#7835)
- `5c8032d70` chore(deps): bump react-i18next from 17.0.10 to 17.1.0 in /web (#7981)
- `0b667ca47` chore(deps): bump i18next from 26.3.6 to 26.5.3 in /web (#7979)
- `4105bd089` chore(deps): bump input-otp from 1.4.2 to 1.5.0 in /web (#7980)

以上为最新 10 条，本次共 14 条；其余 4 条为 `b6b85cae5` chore(deps): bump react-resizable-panels from 4.12.2 to 4.14.6 in /web (#7982)、`1dafb19a4` fix(gemini): preserve top-level imageConfig in chat conversion、`0034b2491` fix(task): wait up to 72 hours before timing out tasks、`a40446041` chore(web): upgrade Bun to 1.4.4 and use bun check。

---

## v1.0.4-ie — 2026-10-23

同步上游 2 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

- `33c61a242` fix: prevent stale values when switching edit drawer records (#7989)
- `30748c655` fix(logging): 隐藏访问日志中的查询凭据 (#7972)

---

## v1.0.3-ie — 2026-10-23

同步上游 12 个提交，无二开改动。合并无冲突，编译与回归测试通过。

**同步上游**

- `7fb6e57a5` fix(channels): preserve advanced custom body passthrough settings (#7971)
- `e9d4d6353` feat(billing): bill OpenAI 12-hour prompt cache writes at their own price (#7965)
- `d0b840e75` fix(relay): keep the client's JSON Schema property order
- `d8e65613c` feat(web): copy model names from usage log details
- `880211326` feat(task): let ordinary channels compete on shared plugin endpoints
- `eb169e74d` feat(pricing): adjust synced upstream prices by the site exchange rate
- `e2321a5a9` chore(web): sort locale keys
- `db80ac8e0` fix(task): let task plugins validate their own requests
- `dbe97beb9` feat(task): pass client request headers to task plugin drivers
- `556517fa5` feat(task): accept multipart bodies on native plugin routes

以上为最新 10 条，本次共 12 条；其余 2 条为 `94cc3ff09` fix(billing): bill Claude cache writes that arrive without a 5m/1h split、`7e6596728` fix(relay): stop cutting off slow clients during large stream events。

---

## v1.0.2-ie — 2026-10-22

首个带 `-ie` 后缀的版本。内容为同步上游 16 个提交，无二开改动。

**同步上游**

- `0c2af9780` fix(relay): finish delivered Responses streams without false disconnect errors (#7956)
- `7d8b91268` fix: upgrade Go to 1.26.10 for CVE-2026-111943 (#7942)
- `4d36f2c52` chore(deps-dev): bump vitest from 4.1.10 to 5.0.4 in /web (#7838)
- `eea05e010` chore(deps): bump github.com/klauspost/compress from 1.18.3 to 1.18.7 (#7931)
- `206e0fa5c` fix(kling): bump plugin to 1.1.1
- `590042682` fix(task): restrict query redirect credentials to allowed hosts
- `120dbeafb` test(security): match the email code only as a whole number
- `f9a0634f8` test(audit): read PostgreSQL index columns in key order
- `55e2ae54d` fix(task): reprice each retry and reserve it before submission
- `fc8f7d6fe` fix: fall back to CNY when DeepSeek USD wallet is zero (#7917)
- `e2a0c4b16` fix(limiter): recover redis script after flush (#6486)
- `a8535fac1` fix(kling): accept single API keys (#7897)
- `783735607` fix(relaykit): keep tool calls when merging adjacent assistant messages for Claude (#7921)
- `afe73b08f` feat: 支持阿里多模态向量化模型 qwen3-vl-embedding (#5179)
- `9a42033a6` fix(task): 修复任务提交与轮询的请求头覆盖 (#7918)
- `603a5ceb5` feat(dashboard): show cache usage alongside total tokens (#7773)

---

## v1.0.1 — 2026-10-22

首个正式编号版本，功能等价于 `rc.50`，另补 README 与更新日志。

**新增**

- 首页：`GET /` 返回静态服务提示页，页脚展示版本号与构建时间
- 构建：新增 `common.BuildTime`，构建时经 `-ldflags` 注入；未注入时回退 Go build info 的 `vcs.time`
- 构建时间注入覆盖 release（Linux/macOS/Windows）、electron、Dockerfile、Dockerfile.dev

**变更**

- 文档：README 精简为二开说明与快速开始，新增本更新日志

---

## 历史版本（预览版）

### v1.0.0-rc.50 — 2026-10-22

- 首页改为静态提示页，通过 `router.ServeHomePage` 在 SPA 之前拦截 `/`，其余路径仍走控制台
- 页脚显示版本与构建时间，版本值做 HTML 转义

### v1.0.0-rc.49 — 2026-10-21

- 令牌用量接口 `user_quota.total_available` 改为「钱包 + 全部活跃套餐额度」，与 `total_available_usd` 口径一致
- 新增 `wallet_total_available` 字段，保留原「仅钱包」口径

### v1.0.0-rc.48 — 2026-10-20

- 降低兑换接口延迟，精简兑换后返回的令牌信息

### v1.0.0-rc.47 — 2026-10-20

- 新增两种兑换入口：令牌兑换、LDAP 兑换
- 用量接口增加钱包与总资产金额
- 新增 MySQL 版编排文件 `docker-compose.mysql.yml`

### v1.0.0-rc.46 — 2026-10-20

- 订阅额度返回折算后的金额字段（`amount_total_usd` / `amount_used_usd` / `amount_available_usd`）

### v1.0.0-rc.45 — 2026-10-20

首个二开版本。

- LDAP 凭据换取 API 令牌：`POST /api/user/ldap-login-return-token`
- 令牌用量查询接口返回订阅额度：`GET /api/usage/token/`
- 管理端新增 LDAP 目录设置面板
- Docker 镜像改为 `handloong/new-api-ie`

---

## 接口文档

- LDAP 登录换 Token、Token 用量查询：[issue #2](https://github.com/handloong/new-api-ie/issues/2)
- CDK 兑换：[issue #3](https://github.com/handloong/new-api-ie/issues/3)
