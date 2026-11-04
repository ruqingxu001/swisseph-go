# Swiss Ephemeris Go

**简体中文** | [English](README.en.md)

Swiss Ephemeris **2.10.03** 的 Go 移植，保留原算法及 AGPL。九个核心模块现在使用同一套原生 Go 实现，运行时只依赖标准库。核心由明确的结构体、切片、模块状态和文件解码组成，已移除 `unsafe`、`modernc.org/libc`、C 指针/内存布局模拟及 `goto`。

**106 个公开 API 均有 Go 接口和原 C 样本对照。** 已实际运行 Flutter 的 sweph 包装层及 App CalcTools 做比较。逐函数结果、容差和未覆盖范围见 [API_COVERAGE.md](API_COVERAGE.md)。

独立模块与业务项目无依赖。仓库：[ruqingxu001/swisseph-go](https://github.com/ruqingxu001/swisseph-go)，模块路径为 `github.com/ruqingxu001/swisseph-go`。当前为移植验证版，尚未接入产品。

## 使用

最低 Go 1.24.2。本轮在 Windows/amd64、Linux/amd64 实测；macOS/Linux/Windows × amd64/arm64 六个目标交叉构建。交叉构建不等于在对应机器运行。

```go
import swiss "github.com/ruqingxu001/swisseph-go"

engine, err := swiss.New(swiss.Config{
    EphemerisPath: "./ephe", // 在本仓库根目录运行；部署时传实际数据目录。
    SiderealMode: swiss.Lahiri,
})
if err != nil { return err }
jd, err := engine.JulianDay(2000, 1, 1, 12, swiss.Gregorian)
if err != nil { return err }
sun, err := engine.CalcUT(jd, swiss.Sun, swiss.SwissEphemeris|swiss.Speed)
if err != nil { return err }
_ = sun.Position[0] // 黄经（度）；Position[3] 为黄经速度（度/日）。
```

有两层接口：

| 接口 | 用途 |
| --- | --- |
| `Engine` | 常用计算：位置、宫位、日期、UTC 转换、日月食；检查错误和文件回退 |
| `Engine.WithSession` → `Session` | 全部 106 个 C API；连续批量计算、固定星、掩星、偕日升降、JPL 等 |

`Session` 方法及缓冲区长度见 [api_generated.go](api_generated.go)，机器可读清单见 [api.json](api.json)。例如：

```go
err = engine.WithSession(func(s *swiss.Session) error {
    name, err := s.GetPlanetName(swiss.Mars, make([]byte, 256))
    if err != nil { return err }
    _ = name
    return nil
})
```

- UT 是 UT1，ET 是 TT。需要严格 UTC→UT1/TT 转换时使用 `UTCToJD`。Flutter 当前出生日期流程直接把 UTC 日历换成 JD，对齐该流程时使用 `JulianDay`。
- `Engine` 默认拒绝缺文件后的算法回退；`AllowFallback: true` 可允许，返回的 `Flags`/`Warning` 保留实际来源。宫位失败仍保留 C 的备用结果。
- `Session` 保留 C 返回码、诊断和回退规则；Go `error` 只表示会话或输入缓冲区错误。调用者还须检查 C 状态码。
- 每次 `Engine` 操作拥有独立状态，可并行计算；`Session` 在回调内复用缓存并按方法加锁。对照连续 Flutter 调用应使用同样的会话与顺序，尤其是历史日期。
- 嵌套 `Engine` 操作使用独立状态；回调结束后原会话失效。调用者负责共享输出切片的同步；不同参数不得依赖 C 指针别名行为。
- 使用可信且计算期间不变的星历文件；路径必须是单个目录，禁止环境变量 `SE_EPHE_PATH` 覆盖配置。

## 验证

```sh
CGO_ENABLED=0 go test ./...

# 在仓库根目录使用随库提供的数据。
export SWISSGO_TEST_EPHE_PATH="$PWD/ephe"
export SWISSGO_TEST_JPL_FILE="$PWD/testdata/jpl/de200.eph"
CGO_ENABLED=0 go test -tags differential,fluttercompare -v ./... -timeout 10m

go test -race -tags differential,fluttercompare ./... -timeout 10m
go vet ./...
```

C 编译器与 Python 3 只用于构建测试参照；race 检测器自身需要 CGO，计算实现仍是 Go。默认 `go vet ./...` 全部开启。在 Windows 上执行 Python 脚本请使用 `py -3`，完整 C 差分命令在 Linux/macOS 运行。

验证覆盖全部 106 个 API / 357 场景、2,260 组 Engine C 对照、400 组真实 JPL、148 组私有模块 C 对照、497 次参数边界与 1,020 次数字前缀解析对照。374 个原算法函数均有执行覆盖，逐函数语句覆盖率见 [CORE_VALIDATION.md](CORE_VALIDATION.md)。既有 1,885 条 Dart/App 参考继续保留；另在 Linux 重新采集并验证 1,335 条真实 Dart 包装层输出。App 的 550 条输出本轮未重新采集，其源码哈希已复核。完整测试、race 和默认 vet 均通过。

仓库已包含 Flutter 使用的 36 个星历及辅助文件（`ephe/`），以及验证用 DE200（`testdata/jpl/de200.eph`）。来源、体积与使用方式见 [EPHEMERIS.md](EPHEMERIS.md)。校验值分别保存在 `testdata/ephemeris.sha256` 和 `testdata/jpl-source.json`。不设置 JPL 环境变量时该项仍会明确跳过。

## 重新采集 Flutter 参考

`testdata/flutter/` 保存本次实际运行的合成输入结果及源文件哈希，不包含用户档案。默认对照这些结果。重新采集需要对应 Flutter 工程及锁定的 sweph 3.2.1+2.10.3：

```sh
python3 tools/run_flutter_reference.py \
  --dart /path/to/flutter/bin/dart \
  --packages /path/to/app/.dart_tool/package_config.json \
  --ephe /path/to/ephe \
  --output testdata/flutter/wrapper.jsonl \
  --app /path/to/app \
  --flutter /path/to/flutter/bin/flutter \
  --app-output testdata/flutter/app.jsonl
```

脚本编译该 Dart 包自己的 C 源码，先确认与本库固定版本一致，再调用真实包装层和 App CalcTools。独立库的构建和普通测试无需 Flutter 工程。

已识别 Flutter 包的 `Hsys.X/Y` 编码错误（58/59，应为 88/89）。对照保留原包装层行为；Go 的 `Engine.Houses` 使用正确编码。用户 App 和包缓存未修改。


本轮 Linux 包装层参考保存在 `testdata/flutter/linux-wrapper.jsonl`，manifest 记录采集条件及保留的 App 参考。可独立选择它复测：

```sh
SWISSGO_FLUTTER_REFERENCE=testdata/flutter/linux-wrapper.jsonl \
  go test -tags fluttercompare ./...
```

## 生成与开源

```sh
python3 tools/generate.py
python3 tools/generate_api.py
python3 tools/build_api_cases.py
# 完整测试和覆盖率采集后生成两份逐函数矩阵：
# go test -tags differential,fluttercompare -v -coverpkg=./internal/ephemeris -coverprofile=core.cover ./...
# go tool cover -func=core.cover > core-functions.txt
python3 tools/write_coverage.py --log /path/to/full-test.log \
  --functions /path/to/core-functions.txt
```

来源、固定提交、转译调整和许可见 [NOTICE.md](NOTICE.md)，贡献流程见 [CONTRIBUTING.md](CONTRIBUTING.md)。核心位于 `internal/ephemeris`，直接人工维护；`tools/generate.py` 仅校验上游哈希，接口生成器不覆盖算法。CI 包含基础测试、完整差分、Flutter 固定参考、race、默认 vet 及六目标构建。

当前属于移植验证版：未穷举全部输入和数据文件，尚未迁移 App 的全部合盘、推运、阿拉伯点及报告业务，也未做生产压测。样本通过不等于所有场景绝对无误。
