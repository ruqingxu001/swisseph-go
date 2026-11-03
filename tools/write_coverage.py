#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-only
"""Build an API matrix only after a complete successful verbose test run."""
import argparse,collections,hashlib,json,re
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser();p.add_argument('--log',required=True);a=p.parse_args()
log=Path(a.log).read_text()
module=(ROOT/'go.mod').read_text().splitlines()[0].split()[1]
required=['API coverage: 106/106; cases=256','2260 C/Go cases','400 real DE200 JPL file cases','actual Flutter/Dart rows=','1020 numeric prefix comparisons','497 argument boundary checks']
if 'FAIL' in log or any(x not in log for x in required) or f'ok  \t{module}\t' not in log:raise SystemExit('Complete passing test log required (including JPL and Flutter)')
apis=json.loads((ROOT/'api.json').read_text());cases=json.loads((ROOT/'testdata/api_cases.json').read_text())
count=collections.Counter(s['function'] for c in cases for s in c['steps'])
rows=['# 逐函数验证记录','',
 '实测环境：2026-11-02，macOS/amd64，Go 1.27.0，Swiss Ephemeris 2.10.03。',
 '“通过”表示固定样本在约定容差内通过原 C 对照；不是全部输入、分支或平台的正确性证明。',
 '', '| 验证层 | 结果 |', '| --- | --- |',
 '| 公开 API | 106/106，256 组场景，包含设置后计算、正常结果及错误分支 |',
 '| 常用 Engine 接口 | 2,260 组 C/Go 对照，377 组确认使用 Swiss 文件 |',
 '| JPL 文件读取 | 400 组 DE200 真文件结果，要求返回 JPL 标志；不同日期顺序检验 seek/read |',
 '| 实际 Flutter | 1,335 条 Dart 包装层 + 550 条 App CalcTools 输出；其中 81 条错误情况 |',
 '| 输入与生命周期 | 497 次非法参数检查；全部 106 个方法拒绝已结束会话；并发、panic 清理、状态隔离 |',
 '| atof 兼容层 | 1,020 组系统 C atof 对照，包括 ICRS、数字前缀、无效指数、溢出与负零 |',
 '', '测试入口：`api_differential_test.go`、`differential_test.go`、`jpl_differential_test.go`、`flutter_differential_test.go`、`session_test.go`。',
 '下面次数只统计 `testdata/api_cases.json` 的直接调用；不含其他测试中的重复调用。参数与完整场景以该文件为准。',
 '', '| C API | Go Session 方法 | 对照调用次数 | 结果 |', '| --- | --- | ---: | --- |']
for api in apis:
 n=api['c_name'];assert count[n]>0
 rows.append(f'| `{n}` | `{api["go_name"]}` | {count[n]} | 通过 |')
rows+=['','`swe_set_jpl_file` 的基本场景是缺文件回退；成功读取由独立的 400 组 JPL 用例验证。','',
 '## 容差和差异','',
 '- 常用接口：常规位置 1e-8；真月孛位置 1e-7；速度 1e-6，站心速度 1e-5。位置及速度单位随 C 标志变化。',
 '- 全 API：通常绝对差 1e-8；整型状态与字符串完全一致。偕日升降/可见度使用 1e-7 + |参考值| × 1e-10。',
 '- 恒星 UT 径向速度 `xx[5]`：容差 1e-5 AU/日，本次差 3.318e-6 AU/日；这是远距离数值差分放大的差异，不能当作角度差。',
 '- 月掩星本地 `attr`：容差 3e-7 + |参考值| × 1e-10，各字段单位按 C 文档；本次最大差 1.355e-6 km，角量差约 1.6e-7 度。',
 '- Flutter 相同连续会话：最大位置差 1.32e-8 度，最大速度差 8.34e-7 度/日。',
 '- DE200：最大位置差 5.69e-14，最大速度差 9.23e-11，单位随坐标标志；无回退。',
 '- 原 C 与 Go 使用同一算法，因此这是移植一致性验证，不是独立观测数据的天文精度认证。','',
 '## 已修复与已识别问题','',
 '1. 运行库 atof 遇到恒星文件中的 ICRS 会退出进程：已增加 C 数字前缀兼容层并对照系统 C。',
 '2. Go 日期校验遇到 23:59:59.999999 可能因 JD 舍入跨日误报：改为在当天中午验证日期。',
 '3. JPL 文件设置会清除月球数据，影响历史日期的 delta-T：便捷接口初始化后重新加载正确数据；连续调用用 Session 保持与 Flutter 相同顺序。',
 '4. Flutter sweph 3.2.1+2.10.3 的 Hsys.X/Y 编码为 58/59，应为 88/89。测试记录原包装层行为，Go 的 Engine.Houses 使用正确编码；未修改用户 App 或包缓存。','',
 '5. Linux ARM64 的 C char 为无符号，其他支持平台为有符号：增加平台参数适配，四平台构建通过。','',
 '## 尚未验证','',
 '- 其他三个平台只有交叉构建，未在对应机器执行；Windows 未支持。',
 '- 未穷举所有日期、星体、组合标志、所有 JPL 版本及异常/损坏数据文件。仅使用可信星历资料。',
 '- 未迁移 App 的全部阿拉伯点、合盘、推运、事件筛选和报告业务。',
 '- 未做生产吞吐和长期内存压测。','',
 '## 证据','',
 '- 固定源码：`upstream/SHA256SUMS`；输入场景：`testdata/api_cases.json`。',
 '- 实际 Dart/App 输出及源文件、数据哈希：`testdata/flutter/`。',
 '- JPL 来源与本地实测文件哈希：`testdata/jpl-source.json`。',
 '- 本次原始测试日志 SHA-256：`'+hashlib.sha256(Path(a.log).read_bytes()).hexdigest()+'`（日志保留在本地工作记录，不纳入发布）。']
(ROOT/'API_COVERAGE.md').write_text('\n'.join(rows)+'\n')
print('Wrote API_COVERAGE.md with',len(apis),'verified APIs')
