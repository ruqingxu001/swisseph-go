# Contributing

本项目按 AGPL-3.0-only 接收修改，保留上游来源、版权及许可声明。

- 直接维护 `internal/ephemeris`；按历法、宫位、JPL、月球、行星、位置、数学/时间、日月食、偕日模块阅读。`*_tables.go` 是只读系数；`State` 仅存每次计算自己的缓存。
- 新计算路径使用原生 Go 类型、切片和结构化控制流；保留公式运算顺序及 binary64 舍入，特别是会被速度差分放大的项。不要引入 `unsafe`、ABI 复制或进程全局可变计算状态。
- `tools/generate.py` 校验固定 C 源码，**不会生成或覆盖算法**。接口以 `upstream/swephexp.h` 为准，运行 `tools/generate_api.py` 更新两层接口。
- 为算法修改添加 C 或真实 Dart 参照，避免用被测 Go 实现生成期望值。`testdata/core_functions.json` 映射全部 374 个原函数，`CORE_VALIDATION.md` 区分函数执行与语句覆盖。
- 保留既有测试阈值；新增场景需要说明单位、数值方法及阈值依据。新增 API 场景的参数名必须与头文件一致，生成器拒绝未识别的覆盖参数。
- 运行 README 中的基础测试、完整 C/Flutter/JPL 差分、race、默认 vet 及跨平台构建。完整验证后用 `tools/write_coverage.py --log ... --functions ...` 更新两份矩阵。
- 测试使用合成日期、位置和星体，不提交个人出生资料、业务配置、凭据或本机路径。数据变化须核对来源并更新哈希及真实参考。
- Windows 运行 Python 使用 `py -3`；完整 C oracle 当前以 Linux/macOS 为执行环境。其他平台的交叉构建不代替实际数值验证。
