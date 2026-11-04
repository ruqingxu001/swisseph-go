#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-only
"""Run the real locked Dart wrapper using its own C sources."""
import argparse,hashlib,json,os,re,subprocess,tempfile
from pathlib import Path
from urllib.parse import urlparse,unquote
ROOT=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser();p.add_argument('--dart',required=True);p.add_argument('--packages',required=True);p.add_argument('--ephe',required=True);p.add_argument('--output',required=True);p.add_argument('--app');p.add_argument('--flutter');p.add_argument('--app-output')
a=p.parse_args();config=Path(a.packages).resolve();packages=json.loads(config.read_text())['packages'];pkg=next(x for x in packages if x['name']=='sweph')
uri=pkg['rootUri'];root=Path(unquote(urlparse(uri).path)) if uri.startswith('file:') else (config.parent/uri).resolve()
version=re.search(r'^version:\s*(\S+)',(root/'pubspec.yaml').read_text(),re.M).group(1)
if version!='3.2.1+2.10.3':raise SystemExit('Unexpected sweph version: '+version)
src=root/'native/sweph/src';names='swedate swehouse swejpl swemmoon swemplan sweph swephlib swecl swehel'.split()
checks={}
for n in names:
 original=(ROOT/f'upstream/{n}.c').read_text();flutter=(src/f'{n}.c').read_text()
 if original!=flutter:raise SystemExit('Flutter C source differs from pinned Go source: '+n)
 checks[n+'.c']=hashlib.sha256((src/f'{n}.c').read_bytes()).hexdigest()
output=Path(a.output).resolve();output.parent.mkdir(parents=True,exist_ok=True)
with tempfile.TemporaryDirectory(prefix='swiss-dart-') as tmp:
 library=Path(tmp)/('libsweph.dylib' if os.uname().sysname=='Darwin' else 'libsweph.so')
 cmd=[os.environ.get('CC','cc'),'-O2','-ffp-contract=off','-fPIC','-shared','-DTLSOFF','-DNO_SWE_GLP','-I'+str(src),*[str(src/(n+'.c')) for n in names],'-lm','-o',str(library)]
 subprocess.run(cmd,check=True)
 with output.open('w') as f:
  subprocess.run([a.dart,'--packages='+str(config),str(ROOT/'validation/flutter_oracle.dart'),str(library.parent/"sweph"),str(Path(a.ephe).resolve())],stdout=f,check=True)
 if a.app:
  if not a.flutter or not a.app_output:p.error('--app needs --flutter and --app-output')
  env=dict(os.environ,SWISSGO_REFERENCE_LIBRARY=str(library.parent/"sweph"),SWISSGO_TEST_EPHE_PATH=str(Path(a.ephe).resolve()),SWISSGO_APP_REFERENCE=str(Path(a.app_output).resolve()))
  subprocess.run([a.flutter,'test','--no-pub',str(ROOT/'validation/flutter_app_test.dart'),'--reporter','expanded'],cwd=Path(a.app).resolve(),env=env,check=True)
# Reference errors can contain local ephemeris paths; publish synthetic data only.
for file in [output]+([Path(a.app_output).resolve()] if a.app else []):
 file.write_text(file.read_text().replace(str(Path(a.ephe).resolve()),'@ephe'))
manifest={'sweph_version':version,'native_source_sha256':checks,
 'fixtures_sha256':{output.name:hashlib.sha256(output.read_bytes()).hexdigest()},
 'dart_source_sha256':{str(f.relative_to(root)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((root/'lib').rglob('*.dart'))},
 'data_sha256':{str(f.relative_to(Path(a.ephe))):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted(Path(a.ephe).rglob('*')) if f.is_file()},
 'runner_sha256':{str(f.relative_to(ROOT)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((ROOT/'validation').glob('*.dart'))}}
if a.app:
 app_output=Path(a.app_output).resolve()
 manifest['fixtures_sha256'][app_output.name]=hashlib.sha256(app_output.read_bytes()).hexdigest()
 app_root=Path(a.app).resolve()
 manifest['app_source_sha256']={str(f.relative_to(app_root)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((app_root/'lib/core/swephx').rglob('*.dart'))}
output.with_suffix('.manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('Dart reference rows:',len(output.read_text().splitlines()))
