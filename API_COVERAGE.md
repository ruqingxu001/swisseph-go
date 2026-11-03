# 逐函数验证记录

实测环境：2026-11-02，macOS/amd64，Go 1.27.0，Swiss Ephemeris 2.10.03。
“通过”表示固定样本在约定容差内通过原 C 对照；不是全部输入、分支或平台的正确性证明。

| 验证层 | 结果 |
| --- | --- |
| 公开 API | 106/106，256 组场景，包含设置后计算、正常结果及错误分支 |
| 常用 Engine 接口 | 2,260 组 C/Go 对照，377 组确认使用 Swiss 文件 |
| JPL 文件读取 | 400 组 DE200 真文件结果，要求返回 JPL 标志；不同日期顺序检验 seek/read |
| 实际 Flutter | 1,335 条 Dart 包装层 + 550 条 App CalcTools 输出；其中 81 条错误情况 |
| 输入与生命周期 | 497 次非法参数检查；全部 106 个方法拒绝已结束会话；并发、panic 清理、状态隔离 |
| atof 兼容层 | 1,020 组系统 C atof 对照，包括 ICRS、数字前缀、无效指数、溢出与负零 |

测试入口：`api_differential_test.go`、`differential_test.go`、`jpl_differential_test.go`、`flutter_differential_test.go`、`session_test.go`。
下面次数只统计 `testdata/api_cases.json` 的直接调用；不含其他测试中的重复调用。参数与完整场景以该文件为准。

| C API | Go Session 方法 | 对照调用次数 | 结果 |
| --- | --- | ---: | --- |
| `swe_heliacal_ut` | `HeliacalUt` | 1 | 通过 |
| `swe_heliacal_pheno_ut` | `HeliacalPhenoUt` | 1 | 通过 |
| `swe_vis_limit_mag` | `VisLimitMag` | 1 | 通过 |
| `swe_heliacal_angle` | `HeliacalAngle` | 1 | 通过 |
| `swe_topo_arcus_visionis` | `TopoArcusVisionis` | 1 | 通过 |
| `swe_set_astro_models` | `SetAstroModels` | 1 | 通过 |
| `swe_get_astro_models` | `GetAstroModels` | 2 | 通过 |
| `swe_version` | `Version` | 1 | 通过 |
| `swe_get_library_path` | `GetLibraryPath` | 1 | 通过 |
| `swe_calc` | `Calc` | 1 | 通过 |
| `swe_calc_ut` | `CalcUt` | 9 | 通过 |
| `swe_calc_pctr` | `CalcPctr` | 1 | 通过 |
| `swe_solcross` | `Solcross` | 1 | 通过 |
| `swe_solcross_ut` | `SolcrossUt` | 1 | 通过 |
| `swe_mooncross` | `Mooncross` | 1 | 通过 |
| `swe_mooncross_ut` | `MooncrossUt` | 1 | 通过 |
| `swe_mooncross_node` | `MooncrossNode` | 1 | 通过 |
| `swe_mooncross_node_ut` | `MooncrossNodeUt` | 1 | 通过 |
| `swe_helio_cross` | `HelioCross` | 1 | 通过 |
| `swe_helio_cross_ut` | `HelioCrossUt` | 1 | 通过 |
| `swe_fixstar` | `Fixstar` | 1 | 通过 |
| `swe_fixstar_ut` | `FixstarUt` | 1 | 通过 |
| `swe_fixstar_mag` | `FixstarMag` | 1 | 通过 |
| `swe_fixstar2` | `Fixstar2` | 1 | 通过 |
| `swe_fixstar2_ut` | `Fixstar2Ut` | 2 | 通过 |
| `swe_fixstar2_mag` | `Fixstar2Mag` | 1 | 通过 |
| `swe_close` | `Close` | 1 | 通过 |
| `swe_set_ephe_path` | `SetEphePath` | 1 | 通过 |
| `swe_set_jpl_file` | `SetJplFile` | 1 | 通过 |
| `swe_get_planet_name` | `GetPlanetName` | 1 | 通过 |
| `swe_set_topo` | `SetTopo` | 1 | 通过 |
| `swe_set_sid_mode` | `SetSidMode` | 1 | 通过 |
| `swe_get_ayanamsa_ex` | `GetAyanamsaEx` | 2 | 通过 |
| `swe_get_ayanamsa_ex_ut` | `GetAyanamsaExUt` | 1 | 通过 |
| `swe_get_ayanamsa` | `GetAyanamsa` | 1 | 通过 |
| `swe_get_ayanamsa_ut` | `GetAyanamsaUt` | 1 | 通过 |
| `swe_get_ayanamsa_name` | `GetAyanamsaName` | 1 | 通过 |
| `swe_get_current_file_data` | `GetCurrentFileData` | 2 | 通过 |
| `swe_date_conversion` | `DateConversion` | 2 | 通过 |
| `swe_julday` | `Julday` | 15 | 通过 |
| `swe_revjul` | `Revjul` | 1 | 通过 |
| `swe_utc_to_jd` | `UtcToJd` | 3 | 通过 |
| `swe_jdet_to_utc` | `JdetToUtc` | 1 | 通过 |
| `swe_jdut1_to_utc` | `Jdut1ToUtc` | 1 | 通过 |
| `swe_utc_time_zone` | `UtcTimeZone` | 1 | 通过 |
| `swe_houses` | `Houses` | 1 | 通过 |
| `swe_houses_ex` | `HousesEx` | 1 | 通过 |
| `swe_houses_ex2` | `HousesEx2` | 4 | 通过 |
| `swe_houses_armc` | `HousesArmc` | 1 | 通过 |
| `swe_houses_armc_ex2` | `HousesArmcEx2` | 1 | 通过 |
| `swe_house_pos` | `HousePos` | 1 | 通过 |
| `swe_house_name` | `HouseName` | 1 | 通过 |
| `swe_gauquelin_sector` | `GauquelinSector` | 1 | 通过 |
| `swe_sol_eclipse_where` | `SolEclipseWhere` | 1 | 通过 |
| `swe_lun_occult_where` | `LunOccultWhere` | 2 | 通过 |
| `swe_sol_eclipse_how` | `SolEclipseHow` | 1 | 通过 |
| `swe_sol_eclipse_when_loc` | `SolEclipseWhenLoc` | 1 | 通过 |
| `swe_lun_occult_when_loc` | `LunOccultWhenLoc` | 1 | 通过 |
| `swe_sol_eclipse_when_glob` | `SolEclipseWhenGlob` | 1 | 通过 |
| `swe_lun_occult_when_glob` | `LunOccultWhenGlob` | 1 | 通过 |
| `swe_lun_eclipse_how` | `LunEclipseHow` | 1 | 通过 |
| `swe_lun_eclipse_when` | `LunEclipseWhen` | 1 | 通过 |
| `swe_lun_eclipse_when_loc` | `LunEclipseWhenLoc` | 1 | 通过 |
| `swe_pheno` | `Pheno` | 1 | 通过 |
| `swe_pheno_ut` | `PhenoUt` | 1 | 通过 |
| `swe_refrac` | `Refrac` | 13 | 通过 |
| `swe_refrac_extended` | `RefracExtended` | 13 | 通过 |
| `swe_set_lapse_rate` | `SetLapseRate` | 1 | 通过 |
| `swe_azalt` | `Azalt` | 2 | 通过 |
| `swe_azalt_rev` | `AzaltRev` | 1 | 通过 |
| `swe_rise_trans_true_hor` | `RiseTransTrueHor` | 1 | 通过 |
| `swe_rise_trans` | `RiseTrans` | 1 | 通过 |
| `swe_nod_aps` | `NodAps` | 1 | 通过 |
| `swe_nod_aps_ut` | `NodApsUt` | 1 | 通过 |
| `swe_get_orbital_elements` | `GetOrbitalElements` | 1 | 通过 |
| `swe_orbit_max_min_true_distance` | `OrbitMaxMinTrueDistance` | 1 | 通过 |
| `swe_deltat` | `Deltat` | 4 | 通过 |
| `swe_deltat_ex` | `DeltatEx` | 7 | 通过 |
| `swe_time_equ` | `TimeEqu` | 4 | 通过 |
| `swe_lmt_to_lat` | `LmtToLat` | 4 | 通过 |
| `swe_lat_to_lmt` | `LatToLmt` | 4 | 通过 |
| `swe_sidtime0` | `Sidtime0` | 4 | 通过 |
| `swe_sidtime` | `Sidtime` | 4 | 通过 |
| `swe_set_interpolate_nut` | `SetInterpolateNut` | 1 | 通过 |
| `swe_cotrans` | `Cotrans` | 1 | 通过 |
| `swe_cotrans_sp` | `CotransSp` | 1 | 通过 |
| `swe_get_tid_acc` | `GetTidAcc` | 2 | 通过 |
| `swe_set_tid_acc` | `SetTidAcc` | 1 | 通过 |
| `swe_set_delta_t_userdef` | `SetDeltaTUserdef` | 1 | 通过 |
| `swe_degnorm` | `Degnorm` | 11 | 通过 |
| `swe_radnorm` | `Radnorm` | 11 | 通过 |
| `swe_rad_midp` | `RadMidp` | 11 | 通过 |
| `swe_deg_midp` | `DegMidp` | 11 | 通过 |
| `swe_split_deg` | `SplitDeg` | 2 | 通过 |
| `swe_csnorm` | `Csnorm` | 1 | 通过 |
| `swe_difcsn` | `Difcsn` | 1 | 通过 |
| `swe_difdegn` | `Difdegn` | 11 | 通过 |
| `swe_difcs2n` | `Difcs2N` | 1 | 通过 |
| `swe_difdeg2n` | `Difdeg2N` | 11 | 通过 |
| `swe_difrad2n` | `Difrad2N` | 11 | 通过 |
| `swe_csroundsec` | `Csroundsec` | 1 | 通过 |
| `swe_d2l` | `D2L` | 11 | 通过 |
| `swe_day_of_week` | `DayOfWeek` | 2 | 通过 |
| `swe_cs2timestr` | `Cs2Timestr` | 1 | 通过 |
| `swe_cs2lonlatstr` | `Cs2Lonlatstr` | 1 | 通过 |
| `swe_cs2degstr` | `Cs2Degstr` | 1 | 通过 |

`swe_set_jpl_file` 的基本场景是缺文件回退；成功读取由独立的 400 组 JPL 用例验证。

## 容差和差异

- 常用接口：常规位置 1e-8；真月孛位置 1e-7；速度 1e-6，站心速度 1e-5。位置及速度单位随 C 标志变化。
- 全 API：通常绝对差 1e-8；整型状态与字符串完全一致。偕日升降/可见度使用 1e-7 + |参考值| × 1e-10。
- 恒星 UT 径向速度 `xx[5]`：容差 1e-5 AU/日，本次差 3.318e-6 AU/日；这是远距离数值差分放大的差异，不能当作角度差。
- 月掩星本地 `attr`：容差 3e-7 + |参考值| × 1e-10，各字段单位按 C 文档；本次最大差 1.355e-6 km，角量差约 1.6e-7 度。
- Flutter 相同连续会话：最大位置差 1.32e-8 度，最大速度差 8.34e-7 度/日。
- DE200：最大位置差 5.69e-14，最大速度差 9.23e-11，单位随坐标标志；无回退。
- 原 C 与 Go 使用同一算法，因此这是移植一致性验证，不是独立观测数据的天文精度认证。

## 已修复与已识别问题

1. 运行库 atof 遇到恒星文件中的 ICRS 会退出进程：已增加 C 数字前缀兼容层并对照系统 C。
2. Go 日期校验遇到 23:59:59.999999 可能因 JD 舍入跨日误报：改为在当天中午验证日期。
3. JPL 文件设置会清除月球数据，影响历史日期的 delta-T：便捷接口初始化后重新加载正确数据；连续调用用 Session 保持与 Flutter 相同顺序。
4. Flutter sweph 3.2.1+2.10.3 的 Hsys.X/Y 编码为 58/59，应为 88/89。测试记录原包装层行为，Go 的 Engine.Houses 使用正确编码；未修改用户 App 或包缓存。

5. Linux ARM64 的 C char 为无符号，其他支持平台为有符号：增加平台参数适配，四平台构建通过。

## 尚未验证

- 其他三个平台只有交叉构建，未在对应机器执行；Windows 未支持。
- 未穷举所有日期、星体、组合标志、所有 JPL 版本及异常/损坏数据文件。仅使用可信星历资料。
- 未迁移 App 的全部阿拉伯点、合盘、推运、事件筛选和报告业务。
- 未做生产吞吐和长期内存压测。

## 证据

- 固定源码：`upstream/SHA256SUMS`；输入场景：`testdata/api_cases.json`。
- 实际 Dart/App 输出及源文件、数据哈希：`testdata/flutter/`。
- JPL 来源与本地实测文件哈希：`testdata/jpl-source.json`。
- 本次原始测试日志 SHA-256：`91dc5bb2f1b43c8a7af28e05fff0cddfe5dacc70dafd90d075ef70a63f0dfb60`（日志保留在本地工作记录，不纳入发布）。
