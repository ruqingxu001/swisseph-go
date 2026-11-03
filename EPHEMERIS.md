# 星历数据

| 目录 | 内容 | 大小 |
| --- | --- | ---: |
| `ephe/` | Flutter 当前使用的 36 个星历及辅助文件，含行星、月球、小行星和恒星目录 | 12,589,953 字节 |
| `testdata/jpl/de200.eph` | JPL DE200，供二进制星历读取对照测试使用 | 42,958,608 字节 |

文件以普通 Git 文件提交，克隆后即可使用，无需 Git LFS。`ephe/` 是本项目已验证的数据集合，不是 Swiss Ephemeris 全部星历库；不同文件覆盖的日期与星体不同。

## 使用与验证

- 常规计算：将 `Config.EphemerisPath` 指向 `ephe/` 的实际路径。
- 全量测试：在仓库根目录执行 README 的测试命令，设置两个星历环境变量。
- Swiss 数据 SHA-256 清单：`testdata/ephemeris.sha256`，路径相对于 `ephe/`。
- DE200 的地址、大小与 SHA-256：`testdata/jpl-source.json`。
- 数据不得在计算期间修改。`.gitattributes` 禁止 Git 转换这些文件的换行，以保持校验一致。

## 来源

Swiss 数据直接复制自本项目原 Flutter 的 `assets/ephe`，与已有校验清单逐字节一致。上游为 [Swiss Ephemeris](https://github.com/aloistr/swisseph)，保留原始目录头部、说明及 [许可声明](upstream/LICENSE)。`ast0/readme.md` 是原有备注，星体映射以编号和正式数据为准。

DE200 来自 [NASA/JPL 官方下载](https://ssd.jpl.nasa.gov/ftp/eph/planets/Linux/de200/lnxm1600p2170.200)，仅将文件名改为 `de200.eph`，内容未改动。相关资料见 [JPL 行星星历说明](https://ssd.jpl.nasa.gov/planets/eph_export.html)。本项目未改写数据提供方适用的使用条款。
