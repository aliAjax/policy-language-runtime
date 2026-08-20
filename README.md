# Policy Language Runtime 10

纯Go策略DSL编译器与决策运行时，提供REST接口以及可嵌入Evaluator。源码按namespace、module、lexer、parser、typecheck、ir、optimizer、runtime、explain、simulation、distribution领域分层，并通过接口注入存储、时钟、签名与函数注册器。

## 启动

```bash
go test ./...
go vet ./...
go run ./cmd/policyd
```

默认监听`:8095`，支持`POLICY_ADDR`、`POLICY_MAX_BODY`、`POLICY_MAX_STEPS`环境覆盖。`/healthz`、`/readyz`、`/metrics`为运维端点。

## REST示例

```bash
curl -s localhost:8095/v1/compile -H 'content-type: application/json' -d '{"source":"return age >= 18;"}'
curl -s localhost:8095/v1/publish -H 'content-type: application/json' -d '{"namespace":"demo","module":"access","version":"1.0.0","channel":"stable","source":"return age >= 18;"}'
curl -s localhost:8095/v1/decide -H 'content-type: application/json' -d '{"namespace":"demo","module":"access","input":{"age":21}}'
curl -s localhost:8095/v1/explain -H 'content-type: application/json' -d '{"namespace":"demo","module":"access","input":{"age":15}}'
```

DSL支持字面量、变量、算术/比较/逻辑运算、短路条件、let、if/else、return及contains、matches、now内置函数。编译阶段限制令牌与嵌套深度，运行阶段支持context取消、步骤预算和未知变量indeterminate结果。

## 持久化与部署

`migrations/`包含策略命名空间、版本、发布通道与样本表迁移；当前开发模式使用内存Store，生产可实现`platform.Store`接入PostgreSQL或对象存储。`deploy/docker-compose.yml`提供依赖示例，`scripts/start.sh`用于启动。

## 验证

建议验证语法定位、类型错误、超大输入拒绝、决策取消、版本发布与差异模拟。服务收到SIGINT/SIGTERM后在5秒内优雅停机，正在执行的决策通过context传播取消。
