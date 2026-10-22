# 更新日志

本文件记录 new-api-ie 二开版本的重要变更，按版本倒序排列。

版本号规则：自 `v1.0.1` 起使用正式编号 `v主.次.修订`；此前的 `v1.0.0-rc.45` ~ `rc.50` 为二开预览版，标签保留。二开起点是 `rc.45`。

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
