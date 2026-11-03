# Contributing

本项目按 AGPL-3.0-only 接收修改，保留上游来源及版权声明。

- 不直接修改 `internal/raw/swe_*.go`。修改生成器后重新生成四个平台，并核对原始源码哈希。
- API 以固定版本 `swephexp.h` 为准；运行 `tools/generate_api.py` 更新接口和清单。
- 新增或调整算法路径时添加原 C 差分场景，不用 Go 自己生成期望值。
- 所有计算测试使用合成日期、位置和星体；不提交个人出生资料、业务配置、密钥或本机绝对路径。
- 普通 CI 不包含未分发的星历文件。发布前按 README 跑全量差分、Flutter 数据对照、JPL 文件及 race，并在目标平台实测。
- 模块路径为 `github.com/ruqingxu001/swisseph-go`。修改公开接口或生成器时同步更新文档与逐函数验证表。
