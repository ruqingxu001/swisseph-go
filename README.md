# Swiss Ephemeris Go

Swiss Ephemeris **2.10.03** 的独立 Go 移植，保留原算法及 AGPL。9 个核心 C 单元已转译；运行时不需要 CGO、libswe 或 C 子进程，使用纯 Go `modernc.org/libc` 和 `unsafe`。这是 C 结构的直接转译，不是重新设计的天文算法。

**106 个公开 API 均有 Go 接口和原 C 样本对照。** 已实际运行 Flutter 的 sweph 包装层及 App CalcTools 做比较。逐函数结果、容差和未覆盖范围见 [API_COVERAGE.md](API_COVERAGE.md)。

独立模块与业务项目无依赖。仓库：[ruqingxu001/swisseph-go](https://github.com/ruqingxu001/swisseph-go)，模块路径为 `github.com/ruqingxu001/swisseph-go`。当前为移植验证版，尚未接入产品。

## 使用

最低 Go 1.24.2；本次在 Go 1.27.0、macOS/amd64 实测。支持 macOS/Linux × amd64/arm64，其他三平台目前只验证交叉构建。

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
- 同一进程内计算串行。`Engine` 每次重置配置；`Session` 保留回调内的缓存和设置。对照连续 Flutter 调用应使用同样的会话与顺序，尤其是历史日期。
- 不在 `WithSession` 回调内调用 `Engine` 或嵌套会话；回调结束后会话失效。不同参数的切片不得依赖 C 指针别名行为。
- 使用可信且计算期间不变的星历文件；路径必须是单个目录，禁止环境变量 `SE_EPHE_PATH` 覆盖配置。

## 验证

```sh
CGO_ENABLED=0 go test ./...

# 在仓库根目录使用随库提供的数据。
export SWISSGO_TEST_EPHE_PATH="$PWD/ephe"
export SWISSGO_TEST_JPL_FILE="$PWD/testdata/jpl/de200.eph"
CGO_ENABLED=0 go test -tags differential,fluttercompare -v ./... -timeout 5m

go test -race -tags differential,fluttercompare ./... -timeout 5m
go vet -unsafeptr=false -unreachable=false -assign=false ./...
```

C 编译器与 Python 3 只用于构建测试参照。race 工具自身需要 CGO，计算实现没有变。普通 `go vet` 对转译代码会报 `unsafe`、不可达语句及自赋值，因此上面仅关闭这三类分析器。

本次通过：106 API / 256 场景、2,260 组常用接口、400 组真 JPL 文件、1,885 条实际 Dart/App 输出、497 次非法参数检查、1,020 次 atof 比较，以及 race。四个平台构建通过；四平台源码及 API 重新生成后与现有文件一致。

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

## 生成与开源

```sh
python3 tools/generate.py
python3 tools/generate_api.py
python3 tools/build_api_cases.py
# 完整测试日志通过后生成逐函数矩阵：
python3 tools/write_coverage.py --log /path/to/full-test.log
```

来源、固定提交、转译调整和许可见 [NOTICE.md](NOTICE.md)，贡献流程见 [CONTRIBUTING.md](CONTRIBUTING.md)。CI 运行基础验证；带星历数据的完整验收使用上面的 `differential,fluttercompare` 命令。

当前属于移植验证版：未穷举全部输入和数据文件，尚未迁移 App 的全部合盘、推运、阿拉伯点及报告业务，也未做生产压测。样本通过不等于所有场景绝对无误。
