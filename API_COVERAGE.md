# 公开 API 验证记录

本轮环境：2026-11-03，Linux/amd64（Go 1.24.2、GCC 11.2.1）、Windows/amd64（Go 1.24.2）。
通过表示固定样本在声明容差内与原 C 或真实 Dart 输出一致，不代表穷举全部输入。

| 验证层 | 通过结果 |
| --- | --- |
| 公开 API | 106/106；357 个场景，含配置、正常结果、错误与多模型 |
| Engine | 2,260 组 C/Go 对照；377 组确认使用 Swiss 文件 |
| JPL | 400 组真实 DE200，要求 JPL 来源标志，覆盖不同日期顺序的 seek/read |
| 私有模块 | 148 组直接 C 对照；完整接口另有间接验证 |
| Flutter | 原 1,335 包装层 + 550 App 行；另采集 1,335 Linux Dart 包装层行 |
| 参数与生命周期 | 497 次非法参数检查；106 方法拒绝已结束会话；并发、panic 清理、隔离 |
| 数字解析 | 1,020 组系统 C atof 对照，含 ICRS、无效指数、溢出、负零与 NaN 类别 |

下表次数只统计 `testdata/api_cases.json` 的直接调用。`swe_set_jpl_file` 基本场景为缺文件回退；成功路径由独立 JPL 测试验证。

| C API | Go Session 方法 | 对照调用次数 | 结果 |
| --- | --- | ---: | --- |
| `swe_heliacal_ut` | `HeliacalUt` | 25 | 通过 |
| `swe_heliacal_pheno_ut` | `HeliacalPhenoUt` | 1 | 通过 |
| `swe_vis_limit_mag` | `VisLimitMag` | 1 | 通过 |
| `swe_heliacal_angle` | `HeliacalAngle` | 1 | 通过 |
| `swe_topo_arcus_visionis` | `TopoArcusVisionis` | 1 | 通过 |
| `swe_set_astro_models` | `SetAstroModels` | 38 | 通过 |
| `swe_get_astro_models` | `GetAstroModels` | 2 | 通过 |
| `swe_version` | `Version` | 1 | 通过 |
| `swe_get_library_path` | `GetLibraryPath` | 1 | 通过 |
| `swe_calc` | `Calc` | 71 | 通过 |
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
| `swe_set_sid_mode` | `SetSidMode` | 5 | 通过 |
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
| `swe_houses_ex2` | `HousesEx2` | 8 | 通过 |
| `swe_houses_armc` | `HousesArmc` | 1 | 通过 |
| `swe_houses_armc_ex2` | `HousesArmcEx2` | 1 | 通过 |
| `swe_house_pos` | `HousePos` | 21 | 通过 |
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
| `swe_rise_trans` | `RiseTrans` | 6 | 通过 |
| `swe_nod_aps` | `NodAps` | 1 | 通过 |
| `swe_nod_aps_ut` | `NodApsUt` | 1 | 通过 |
| `swe_get_orbital_elements` | `GetOrbitalElements` | 1 | 通过 |
| `swe_orbit_max_min_true_distance` | `OrbitMaxMinTrueDistance` | 4 | 通过 |
| `swe_deltat` | `Deltat` | 4 | 通过 |
| `swe_deltat_ex` | `DeltatEx` | 22 | 通过 |
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
| `swe_split_deg` | `SplitDeg` | 3 | 通过 |
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

## 数值标准

- 原有 256 API 场景比较规则不变：一般绝对差 1e-8，整型状态及字符串精确一致。
- 新增恒星投影及附加星体场景的速度字段沿用 Engine 的 1e-6 标准，坐标仍为 1e-8。新增速度最大差 1.672e-7；旧机械版本在新虚星场景也超过 1e-8。数值差分会放大微小位置舍入，不保证逐位一致。
- Engine：一般位置 1e-8、真月孛位置 1e-7、速度 1e-6、站心速度 1e-5；单位由 C 标志决定。
- 恒星 UT 径向速度 xx[5]：1e-5 AU/日，实测最大差 3.318e-6 AU/日，不能当作角度误差。
- 月掩星本地 attr：3e-7 + |参考值| × 1e-10；最大差 1.355e-6 km，角量差约 1.6e-7 度。
- 偕日/可见度：1e-7 + |参考值| × 1e-10。
- 原 Flutter 参考最大角度差 1.32e-8 度、角速度差 8.34e-7 度/日。新 Linux 包装层参考与 Windows Go 最大角度差 1.49e-9 度、角速度差 8.34e-7 度/日。
- DE200 最大位置差 5.69e-14、速度差 9.23e-11，单位随标志；无回退。
- NaN 符号/载荷不是 C atof 保证的结果，只比较类别；有限值与负零仍严格比较。

## Flutter 参考与范围

Linux 上以 Dart 3.5.1 重新运行未修改的 sweph 3.2.1+2.10.3，C/Dart 源码哈希匹配。
`linux-wrapper.manifest.json` 记录采集条件。App 的 550 行本轮未重新采集，已核对对应 App 源码哈希。
sweph Hsys.X/Y 仍编码为 58/59，应为 88/89；测试保留其行为，Engine 使用正确编码。

未穷举所有日期、星体、标志组合、JPL 版本及异常文件。macOS/arm64、macOS/amd64、Linux/arm64、Windows/arm64 只有交叉构建。
未迁移 App 的合盘、推运、阿拉伯点及报告业务，未完成生产吞吐或长期内存测试。

## 证据

- 固定源码和数据：`upstream/SHA256SUMS`、`testdata/ephemeris.sha256`、`testdata/jpl-source.json`。
- C/Go 函数映射：`testdata/core_functions.json`；真实 Dart/App 输出及来源哈希：`testdata/flutter/`。
- 完整日志 SHA-256：`7bddb6bdf8685f231edd2f5a4646aab28c3e6b49cd8af8a1e8617537eca6fa4b`。
- 逐函数覆盖率输出 SHA-256：`42a9d7497a638764d042964ae89509adf6ebbc01f7c944888ccc845471e17846`。
- 原始日志保存在本地工作记录；CI 可重新运行并上传 `core.cover`。
