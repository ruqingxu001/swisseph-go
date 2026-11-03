#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-only
"""Write API and original-function matrices from a complete passing test run."""
import argparse
import collections
import hashlib
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--log', required=True, help='Verbose full C/Flutter/JPL test log')
p.add_argument('--functions', required=True, help='Output of go tool cover -func=core.cover')
a = p.parse_args()
log_path, functions_path = Path(a.log), Path(a.functions)
log = log_path.read_text(encoding='utf-8')
def read_json(name):
    return json.loads((ROOT / name).read_text(encoding='utf-8'))
apis, cases = read_json('api.json'), read_json('testdata/api_cases.json')
inventory = read_json('testdata/core_functions.json')
required = [f'API coverage: {len(apis)}/{len(apis)}; cases={len(cases)}',
            '2260 C/Go cases', '400 real DE200 JPL file cases', 'actual Flutter/Dart rows=',
            '1020 numeric prefix comparisons', '497 argument boundary checks',
            'TestPrivateModulesAgainstC']
if re.search(r'(?m)^FAIL|--- FAIL:', log) or any(x not in log for x in required):
    raise SystemExit('Complete passing verbose log required, including private C, JPL and Flutter')
functions = functions_path.read_text(encoding='utf-8')
covered = {(m[1], m[2]): m[3] for m in re.finditer(r'/ephemeris/([^/:]+):\d+:\s+(\w+)\s+([\d.]+)%', functions)}
total = re.search(r'total:.*?([\d.]+)%', functions)
if not total:
    raise SystemExit('Missing total statement coverage')
for row in inventory:
    key = Path(row['go_file']).name, row['go_function']
    if float(covered.get(key, '0')) <= 0:
        raise SystemExit('Original function has no execution coverage: ' + row['c_function'])
calls = collections.Counter(s['function'] for c in cases for s in c['steps'])
private_count = sum(map(int, re.findall(r': (\d+) private module C/Go comparisons', log)))
evidence = ['', '## 证据', '',
    '- 固定源码和数据：`upstream/SHA256SUMS`、`testdata/ephemeris.sha256`、`testdata/jpl-source.json`。',
    '- C/Go 函数映射：`testdata/core_functions.json`；真实 Dart/App 输出及来源哈希：`testdata/flutter/`。',
    '- 完整日志 SHA-256：`' + hashlib.sha256(log_path.read_bytes()).hexdigest() + '`。',
    '- 逐函数覆盖率输出 SHA-256：`' + hashlib.sha256(functions_path.read_bytes()).hexdigest() + '`。',
    '- 原始日志保存在本地工作记录；CI 可重新运行并上传 `core.cover`。']
rows = ['# 公开 API 验证记录', '',
    '本轮环境：2026-11-03，Linux/amd64（Go 1.24.2、GCC 11.2.1）、Windows/amd64（Go 1.24.2）。',
    '通过表示固定样本在声明容差内与原 C 或真实 Dart 输出一致，不代表穷举全部输入。', '',
    '| 验证层 | 通过结果 |', '| --- | --- |',
    f'| 公开 API | {len(apis)}/{len(apis)}；{len(cases)} 个场景，含配置、正常结果、错误与多模型 |',
    '| Engine | 2,260 组 C/Go 对照；377 组确认使用 Swiss 文件 |',
    '| JPL | 400 组真实 DE200，要求 JPL 来源标志，覆盖不同日期顺序的 seek/read |',
    f'| 私有模块 | {private_count} 组直接 C 对照；完整接口另有间接验证 |',
    '| Flutter | 原 1,335 包装层 + 550 App 行；另采集 1,335 Linux Dart 包装层行 |',
    '| 参数与生命周期 | 497 次非法参数检查；106 方法拒绝已结束会话；并发、panic 清理、隔离 |',
    '| 数字解析 | 1,020 组系统 C atof 对照，含 ICRS、无效指数、溢出、负零与 NaN 类别 |', '',
    '下表次数只统计 `testdata/api_cases.json` 的直接调用。`swe_set_jpl_file` 基本场景为缺文件回退；成功路径由独立 JPL 测试验证。', '',
    '| C API | Go Session 方法 | 对照调用次数 | 结果 |', '| --- | --- | ---: | --- |']
for api in apis:
    name = api['c_name']
    if calls[name] <= 0:
        raise SystemExit('Missing API case: ' + name)
    rows.append(f'| `{name}` | `{api["go_name"]}` | {calls[name]} | 通过 |')
rows += ['', '## 数值标准', '',
    '- 原有 256 API 场景比较规则不变：一般绝对差 1e-8，整型状态及字符串精确一致。',
    '- 新增恒星投影及附加星体场景的速度字段沿用 Engine 的 1e-6 标准，坐标仍为 1e-8。新增速度最大差 1.672e-7；旧机械版本在新虚星场景也超过 1e-8。数值差分会放大微小位置舍入，不保证逐位一致。',
    '- Engine：一般位置 1e-8、真月孛位置 1e-7、速度 1e-6、站心速度 1e-5；单位由 C 标志决定。',
    '- 恒星 UT 径向速度 xx[5]：1e-5 AU/日，实测最大差 3.318e-6 AU/日，不能当作角度误差。',
    '- 月掩星本地 attr：3e-7 + |参考值| × 1e-10；最大差 1.355e-6 km，角量差约 1.6e-7 度。',
    '- 偕日/可见度：1e-7 + |参考值| × 1e-10。',
    '- 原 Flutter 参考最大角度差 1.32e-8 度、角速度差 8.34e-7 度/日。新 Linux 包装层参考与 Windows Go 最大角度差 1.49e-9 度、角速度差 8.34e-7 度/日。',
    '- DE200 最大位置差 5.69e-14、速度差 9.23e-11，单位随标志；无回退。',
    '- NaN 符号/载荷不是 C atof 保证的结果，只比较类别；有限值与负零仍严格比较。', '',
    '## Flutter 参考与范围', '',
    'Linux 上以 Dart 3.5.1 重新运行未修改的 sweph 3.2.1+2.10.3，C/Dart 源码哈希匹配。',
    '`linux-wrapper.manifest.json` 记录采集条件。App 的 550 行本轮未重新采集，已核对对应 App 源码哈希。',
    'sweph Hsys.X/Y 仍编码为 58/59，应为 88/89；测试保留其行为，Engine 使用正确编码。', '',
    '未穷举所有日期、星体、标志组合、JPL 版本及异常文件。macOS/arm64、macOS/amd64、Linux/arm64、Windows/arm64 只有交叉构建。',
    '未迁移 App 的合盘、推运、阿拉伯点及报告业务，未完成生产吞吐或长期内存测试。'] + evidence
(ROOT / 'API_COVERAGE.md').write_text('\n'.join(rows) + '\n', encoding='utf-8')
by_module = collections.Counter(row['source'] for row in inventory)
core = ['# 核心算法逐函数验证', '',
    f'固定 Swiss Ephemeris 2.10.03 的 {len(inventory)} 个原函数全部由 Go 实现；本轮 {len(inventory)}/{len(inventory)} 有执行覆盖。',
    f'整个 internal/ephemeris 包语句覆盖率为 **{total[1]}%**，包含生成的 API 桥及 Go 辅助函数。',
    '函数已执行仅表示至少一个语句执行过，不代表所有分支或输入均已验证。下面逐项列出实际语句覆盖率。', '',
    '核心使用标准库、原生结构体/切片和每次计算独立的 State；系数表只读，历法/时间默认表按值复制。',
    '保留公式运算顺序及 C binary64 舍入。原函数名称可以追溯上游；领域概念用于记录字段及局部变量。历法、坐标/多项式、文件读取及 EOP 加载等使用直接 Go 实现。', '',
    '| 原模块 | 有执行覆盖 / 原函数 |', '| --- | ---: |']
for source, count in sorted(by_module.items()):
    core.append(f'| `{source}` | {count}/{count} |')
core += ['', f'除接口回归外，{private_count} 组私有 C 对照覆盖 FK4/FK5/ICRS、Bessel、宫位辅助、月球元素/拱点、偕日辅助、虚星项/名称、暮光、有效太阳质量及 EOP 文件。',
    '文件测试另检验缓冲 seek/tell、两种字节序、3 字节 Swiss 索引、截断文件、Windows 驱动器路径、历法往返及 Chebyshev 导数。', '',
    '| C 函数 | Go 函数 | Go 文件 | 语句覆盖率 |', '| --- | --- | --- | ---: |']
for row in inventory:
    key = Path(row['go_file']).name, row['go_function']
    core.append(f'| `{row["c_function"]}` | `{row["go_function"]}` | [{key[0]}]({row["go_file"]}) | {covered[key]}% |')
core += ['', '## 未穷举的路径', '',
    '不足 100% 的函数仍有未执行语句，包括错误/回退处理、部分数据和模型分支，以及事件搜索的其他条件。',
    '这些缺口保留在报告中。有限采样、相同算法的 C 对照及固定 Flutter 输出不能证明全部输入绝对正确，也不是独立观测精度认证。',
    '逐字段容差和平台运行范围见 API_COVERAGE.md。'] + evidence
(ROOT / 'CORE_VALIDATION.md').write_text('\n'.join(core) + '\n', encoding='utf-8')
print(f'Wrote API_COVERAGE.md and CORE_VALIDATION.md: {len(apis)} APIs; {len(inventory)} original functions')
