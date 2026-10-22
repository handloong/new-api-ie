# new-api-ie

基于 [new-api](https://github.com/QuantumNous/new-api) 的二开版本，面向企业内部接入场景：用 LDAP 账号直接换取 API Token，配额由套餐与钱包共同承载。

## 相比上游的改动

- **LDAP 登录换令牌** — `POST /api/user/ldap-login-return-token`，OA 账号密码换 API Token，首次登录自动建号
- **令牌用量查询** — `GET /api/usage/token/`，返回钱包额度、套餐额度与折算金额
- **CDK 兑换** — `POST /api/user/topup-by-token`，凭兑换码充值
- **首页提示页** — `GET /` 返回静态状态页，页脚显示版本与构建时间
- 管理端 LDAP 设置面板、MySQL 编排文件、Docker 镜像改用 `handloong/new-api-ie`

接口细节见 [issue #2](https://github.com/handloong/new-api-ie/issues/2)、[issue #3](https://github.com/handloong/new-api-ie/issues/3)。

## 快速开始

```bash
docker run -d --name new-api-ie \
  -p 3000:3000 \
  -v $(pwd)/data:/data \
  handloong/new-api-ie:latest
```

启动后访问 `http://localhost:3000`，控制台登录入口是 `/sign-in`。

也可用仓库内的编排文件：

```bash
docker compose up -d                              # SQLite
docker compose -f docker-compose.mysql.yml up -d  # MySQL
```

## 更新日志

见 [CHANGELOG.md](./CHANGELOG.md)。版本号自 `v1.0.1` 起使用正式编号，早期 `v1.0.0-rc.45` ~ `rc.50` 为预览版，标签保留可查。

## 许可证

[AGPL-3.0](./LICENSE)。上游版权归 QuantumNous，二开部分归 Handloong。
