#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-only
"""Maintain readable success and stateful cases for every public API."""
import json
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
apis=json.loads((ROOT/'api.json').read_text())
byname={a['c_name']:a for a in apis}
scalar={
 'tjdstart_ut':2460310.5,'tjd_start':2460310.5,'tjd_ut':2451545.0,'tjd':2451545.0,
 'tjd_et':2451545.0,'tjdut':2451545.0,'t_ut':2451545.0,'jd_et':2451545.0,'jd_ut':2451545.0,'jd':2451545.0,
 'tjd_lmt':2451545.0,'tjd_lat':2451545.0,'ipl':4,'iplctr':5,'iflag':258,'ifl':2,'flags':258,'flag':258,'epheflag':2,
 'geolat':31.2304,'geolon':121.4737,'geoalt':20,'hsys':ord('P'),'armc':123.45,'eps':23.4392911,'nut':0.003,
 'x2cross':120.0,'dir':1,'backward':0,'ifltype':0,'method':0,'imeth':0,
 'y':2024,'year':2024,'iyear':2024,'m':2,'month':2,'imonth':2,'d':29,'day':29,'iday':29,
 'utime':12.345,'hour':12.345,'gregflag':1,'c':ord('g'),'ihour':12,'imin':34,'dsec':56.789,'d_timezone':5.5,
 'sid_mode':1,'isidmode':1,'t0':2451545.0,'ayan_t0':24.0,'ifno':0,
 'atpress':1013.25,'attemp':15,'lapse_rate':0.0065,'inalt':5.0,'calc_flag':0,'rsmi':1,'horhgt':0.5,
 'TypeEvent':1,'helflag':2,'mag':1.0,'azi_obj':100.0,'alt_obj':5.0,'azi_sun':110.0,'azi_moon':80.0,'alt_moon':10.0,
 'do_interpolate':1,'t_acc':-25.8,'dt':0.0008,'ddeg':123.456789,'roundflag':0,
 'x':-12.345678,'x1':350.0,'x0':10.0,'p':-123456,'p1':123456,'p2':789012,
 't':12345678,'sep':ord(':'),'suppressZero':0,'pchar':ord('E'),'mchar':ord('W'),
}
arrays={'geopos':[8.55,47.37,408],'dgeo':[8.55,47.37,408],'datm':[1013.25,15,40,0.25],'dobs':[36,1,1,1,0,0],
 'xpin':[123.45,2.1],'xpo':[123.45,2.1,1.234,0.987,0.012,0.003],'xin':[123.45,2.1,1.234]}
def step(name,overrides=None):
 a=byname[name];args={}
 for p in a['params']:
  n=p['name'];typ=p['gotype']
  if typ=='string':args[n]='@ephe' if n=='path' else 'de431.eph'
  elif typ=='[]byte':args[n]={'star':'Spica','starname':'','ObjectName':'Venus','samod':'5,9,9,4,3,0,0,4'}.get(n,'')
  elif typ in ['[]float64','[]int32']:args[n]=(arrays.get(n,[]) + [0]*p['size'])[:p['size']]
  else:
   if n not in scalar:raise ValueError((name,n))
   args[n]=int(scalar[n]) if typ in ['int32','int8'] else scalar[n]
 if overrides and not overrides.keys() <= args.keys():
  raise ValueError((name, 'unknown parameters', overrides.keys() - args.keys()))
 args.update(overrides or {})
 return {'function':name,'args':args}

cases=[]
for a in apis:
 name=a['c_name'];main=step(name);steps=[main]
 if name.startswith('swe_lun_occult'):main['args']['ipl']=3
 if name in ['swe_sol_eclipse_where','swe_sol_eclipse_how']:main['args']['tjd']=2460409.262
 if name=='swe_lun_eclipse_how':main['args']['tjd_ut']=2459891.958
 if name=='swe_get_current_file_data':steps=[step('swe_calc_ut'),main]
 if name=='swe_close':steps=[step('swe_calc_ut'),main,step('swe_get_current_file_data')]
 if name=='swe_set_ephe_path':steps+=[step('swe_calc_ut')]
 if name=='swe_set_jpl_file':steps+=[step('swe_calc_ut',{'iflag':257})] # Missing-file fallback is separately labelled.
 if name=='swe_set_topo':steps+=[step('swe_calc_ut',{'iflag':258|32768,'ipl':1})]
 if name=='swe_set_sid_mode':steps+=[step('swe_calc_ut',{'iflag':258|65536}),step('swe_get_ayanamsa_ex')]
 if name=='swe_set_astro_models':steps+=[step('swe_get_astro_models'),step('swe_deltat_ex')]
 if name=='swe_set_lapse_rate':steps+=[step('swe_azalt')]
 if name=='swe_set_interpolate_nut':steps+=[step('swe_calc_ut',{'ipl':-1})]
 if name=='swe_set_tid_acc':steps+=[step('swe_get_tid_acc'),step('swe_deltat_ex',{'tjd':2415020.5})]
 if name=='swe_set_delta_t_userdef':steps+=[step('swe_deltat_ex')]
 cases.append({'id':name+'/basic','steps':steps})
# Additional boundary inputs preserve success/error distinction in the report.
for name,overrides,label in [
 ('swe_calc_ut',{'ipl':1000000000},'invalid_body'),
 ('swe_utc_to_jd',{'iyear':2016,'imonth':12,'iday':31,'ihour':23,'imin':59,'dsec':60},'leap_second'),
 ('swe_utc_to_jd',{'iyear':2023,'imonth':2,'iday':29},'invalid_date'),
 ('swe_date_conversion',{'y':2023,'m':2,'d':29},'invalid_date'),
 ('swe_fixstar2_ut',{'star':'not_a_real_star_abcdef'},'missing_star'),
 ('swe_houses_ex2',{'geolat':70},'polar_error'),
 ('swe_houses_ex2',{'hsys':ord('G')},'gauquelin_36'),
 ('swe_houses_ex2',{'iflag':2|65536},'lahiri'),
 ('swe_split_deg',{'ddeg':359.999999,'roundflag':1},'round_boundary'),
 ('swe_day_of_week',{'jd':2299160.5},'calendar_boundary'),
]: cases.append({'id':name+'/'+label,'steps':[step(name,overrides)]})
# Coordinate, calendar and cyclic-angle boundaries are deterministic and cheap.
for value in [-1080.1,-360.0,-0.000001,0.0,0.000001,179.999999,180.0,359.999999,360.0,1080.1]:
 for name in ['swe_degnorm','swe_radnorm','swe_d2l']:
  cases.append({'id':name+'/boundary_'+str(value),'steps':[step(name,{'x':int(value) if name=='swe_d2l' else value})]})
 for name in ['swe_deg_midp','swe_rad_midp','swe_difdegn','swe_difdeg2n','swe_difrad2n']:
  keys=('x1','x0') if 'midp' in name else ('p1','p2')
  cases.append({'id':name+'/boundary_'+str(value),'steps':[step(name,{keys[0]:value,keys[1]:-12.345})]})
for year in [-4000,-1,0,1582,1900,2000,2100]:
 for calendar in [0,1]:
  cases.append({'id':'swe_julday/year_'+str(year)+'_'+str(calendar),'steps':[step('swe_julday',{'year':year,'gregflag':calendar})]})
for altitude in [-5.0,-0.1,0.0,5.0,45.0,89.9]:
 for direction in [0,1]:
  for name in ['swe_refrac','swe_refrac_extended']:
   cases.append({'id':name+'/altitude_'+str(altitude)+'_'+str(direction),'steps':[step(name,{'inalt':altitude,'calc_flag':direction})]})
for jd in [2415020.5,2451545.0,2460310.5]:
 for name in ['swe_sidtime','swe_sidtime0','swe_deltat','swe_deltat_ex','swe_time_equ','swe_lmt_to_lat','swe_lat_to_lmt']:
  call=step(name)
  for key in list(call['args']):
   if key.startswith('tjd') and not isinstance(call['args'][key],list):call['args'][key]=jd
  cases.append({'id':name+'/time_'+str(jd),'steps':[call]})
# Real Venus occultation found by the native global search, then geometry there.
cases.append({'id':'swe_lun_occult_where/at_event','steps':[step('swe_lun_occult_where',{'tjd':2460408.1792892623,'ipl':3})]})
# Exercise every supported precession, nutation and delta-T model rather than
# relying only on the defaults used by the application.
for model in range(1,12):
 for jd in [1500000.5,2458849.5]:
  cases.append({'id':f'model/precession_{model}_{jd}','steps':[
   step('swe_set_astro_models',{'samod':f'5,{model},{model},4,3,0,0,4'}),
   step('swe_calc',{'tjd':jd,'iflag':4|256,'ipl':4}),
   step('swe_calc',{'tjd':jd,'iflag':4|256,'ipl':-1})]})
for model in range(1,6):
 for jd in [1000000.5,2400000.5,2458849.5]:
  cases.append({'id':f'model/nutation_deltaT_{model}_{jd}','steps':[
   step('swe_set_astro_models',{'samod':f'{model},9,9,{model},3,0,0,4'}),
   step('swe_deltat_ex',{'tjd':jd}),step('swe_calc',{'tjd':jd,'ipl':-1,'iflag':4|256})]})
for sidereal_mode in [1|256,1|512,255|256,255|512]:
 cases.append({'id':f'sidereal/projection_{sidereal_mode}','velocity_tolerance':1e-6,'steps':[
  step('swe_set_sid_mode',{'sid_mode':sidereal_mode}),
  step('swe_calc',{'iflag':4|256|65536}),step('swe_houses_ex2',{'iflag':65536})]})
for body in [21,22,40,47,55,10016,10433]:
 cases.append({'id':f'positions/additional_body_{body}','velocity_tolerance':1e-6,'steps':[step('swe_calc',{'ipl':body,'iflag':4|256})]})
for system in ['U','I','i','G']:
 for latitude in [-80,-31.2,0,31.2,80]:
  cases.append({'id':f'houses/position_{system}_{latitude}','steps':[step('swe_house_pos',{'hsys':ord(system),'geolat':latitude})]})
for mode in [4,8,1|1024,2|2048,1|4096]:
 cases.append({'id':f'rise/mode_{mode}','steps':[step('swe_rise_trans',{'rsmi':mode})]})
for flags in [2|65536,2|131072,2|262144,2|524288]:
 for obj in ['Venus','Moon','Sirius']:
  cases.append({'id':f'heliacal/arc_{flags}_{obj}','steps':[step('swe_heliacal_ut',{'iflag':flags|2048,'ObjectName':obj,'TypeEvent':3 if obj=='Moon' else 1})]})
for event in [1,2,3,4,5,6]:
 for obj in ['Moon','Sirius']:
  cases.append({'id':f'heliacal/visibility_{event}_{obj}','steps':[step('swe_heliacal_ut',{'iflag':2|2048,'ObjectName':obj,'TypeEvent':event})]})
for flags in [2,2|8,2|16384]:
 cases.append({'id':f'orbit/distance_flags_{flags}','steps':[step('swe_orbit_max_min_true_distance',{'iflag':flags})]})
cases.append({'id':'degrees/nakshatra','steps':[step('swe_split_deg',{'roundflag':1024})]})
(ROOT/'testdata/api_cases.json').write_text(json.dumps(cases,indent=2)+'\n')
print('Wrote',len(cases),'cases')
