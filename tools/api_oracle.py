#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-only
"""Generate a native-C oracle from the checked-in public API cases."""
import json,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
apis={a['c_name']:a for a in json.loads((ROOT/'api.json').read_text())}
cases=json.loads((ROOT/'testdata/api_cases.json').read_text())
lines=['''/* SPDX-License-Identifier: AGPL-3.0-only */
#include "swephexp.h"
#include <stdio.h>
#include <math.h>
static void text(const char *s) {putchar('"');if(s)for(;*s;s++){unsigned char c=*s;if(c=='"'||c=='\\\\'){putchar('\\\\');putchar(c);}else if(c<32)printf("\\\\u%04x",c);else putchar(c);}putchar('"');}
static void number(double x){if(isfinite(x))printf("%.17g",x);else{fprintf(stderr,"non-finite reference result\\n");exit(3);}}
int main(int argc,char **argv){if(argc!=2)return 2;
''']
for case in cases:
 lines+=['{','swe_close();swe_set_delta_t_userdef(SE_DELTAT_AUTOMATIC);swe_set_interpolate_nut(0);swe_set_lapse_rate(0.0065);swe_set_ephe_path(argv[1]);swe_set_jpl_file("de431.eph");swe_set_ephe_path(argv[1]);swe_set_sid_mode(0,0,0);swe_set_topo(0,0,0);']
 for idx,step in enumerate(case['steps']):
  a=apis[step['function']];expr=[];lines.append('{')
  for p in a['params']:
   n=p['name'];v=step['args'][n];typ=p['gotype']
   if typ=='string':expr.append('argv[1]' if v=='@ephe' else json.dumps(v));continue
   if p['pointer']:
    size=p['size'];ctype={'[]byte':'char','[]float64':'double','[]int32':'int32'}[typ]
    init=json.dumps(v) if typ=='[]byte' else '{'+','.join(str(i) for i in v)+'}'
    lines.append(f'{ctype} {n}[{size}]={init};');expr.append(n)
   else:expr.append(str(v))
  call=f"{a['c_name']}({','.join(expr)})";rtype=a['return']
  if rtype=='void':lines.append(call+';')
  else:lines.append({'string':'const char *','float64':'double','int32':'int32'}[rtype]+' result='+call+';')
  lines.append('printf("{\\"id\\":");text('+json.dumps(case['id'])+');printf(",\\"step\\":'+str(idx)+',\\"function\\":");text('+json.dumps(a['c_name'])+');printf(",\\"result\\":");')
  lines.append('printf("null");' if rtype=='void' else ('text(result);' if rtype=='string' else 'number(result);'))
  lines.append('printf(",\\"outputs\\":{");')
  sep=''
  for p in a['params']:
   if not p['pointer'] or p['gotype']=='string':continue
   n=p['name'];lines.append('printf('+json.dumps(sep)+');text('+json.dumps(n)+');putchar(\':\');');sep=','
   if p['gotype']=='[]byte':lines.append('text('+n+');')
   else:lines.append(f'putchar(\'[\');for(int i=0;i<{p["size"]};i++){{if(i)putchar(\',\');number({n}[i]);}}putchar(\']\');')
  lines.append('puts("}}");\n}')
 lines.append('swe_close();\n}')
lines.append('return 0;}')
Path(sys.argv[1]).write_text('\n'.join(lines)+'\n')
