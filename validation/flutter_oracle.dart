// SPDX-License-Identifier: AGPL-3.0-only
// Runs the actual sweph package selected by --packages. No copied formulas.
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:sweph/sweph.dart';

class Files implements AssetLoader {
  Files(this.path);
  final String path;
  Future<Uint8List> load(String assetPath) => File('$path/$assetPath').readAsBytes();
}
List<double> coords(CoordinatesWithSpeed p) => [p.longitude,p.latitude,p.distance,p.speedInLongitude,p.speedInLatitude,p.speedInDistance];
void emit(Map<String,dynamic> base,dynamic Function() fn) {
  try {base['value']=fn();base['error']=false;} catch (e) {base['error']=true;base['message']=e.toString();}
  stdout.writeln(jsonEncode(base));
}
Future<void> main(List<String> args) async {
  if(args.length!=2) throw ArgumentError('native library and ephemeris directory required');
  await Sweph.init(modulePath:args[0],epheFilesPath:args[1],epheAssets:['sepl_18.se1'],assetLoader:Files(args[1]));
  Sweph.swe_set_sid_mode(SiderealMode.SE_SIDM_LAHIRI,SiderealModeFlag.SE_SIDBIT_NONE,0,0);
  stdout.writeln(jsonEncode({'kind':'metadata','version':Sweph.swe_version()}));
  final dates=[1500000.5,2299160.5,2415020.5,2447892.5,2451545.0,2458849.5,2460409.25,2461346.5,2462502.5];
  final bodies=[for(var i=0;i<=20;i++)i,10016,10433,for(var i=40;i<=47;i++)i];
  for(final jd in dates) for(final body in bodies) for(final sidereal in [false,true]) {
    final flags=258|(sidereal?65536:0);
    emit({'kind':'calc','jd':jd,'body':body,'flags':flags},()=>coords(Sweph.swe_calc_ut(jd,HeavenlyBody(body),SwephFlag(flags))));
  }
  for(final jd in [2447892.5,2451545.0,2460409.25]) for(final hsys in Hsys.values) for(final lat in [0.0,31.2304,-33.8688,70.0,-70.0]) for(final sidereal in [false,true]) {
    final flags=2|(sidereal?65536:0);
    emit({'kind':'houses','jd':jd,'lat':lat,'lon':121.4737,'system':hsys.value,'systemName':hsys.name,'flags':flags},(){
      final h=Sweph.swe_houses_ex2(jd,SwephFlag(flags),lat,121.4737,hsys);
      return {'cusps':h.cusps,'angles':h.ascmc,'cuspSpeeds':h.cuspsSpeed,'angleSpeeds':h.ascmcSpeed};
    });
  }
  for(final year in [-4000,-1,0,1582,1900,2000,2024,2100]) {
    emit({'kind':'julday','year':year},()=>Sweph.swe_julday(year,3,1,12.345,CalendarType.SE_GREG_CAL));
  }
  for(final jd in dates){emit({'kind':'revjul','jd':jd},()=>Sweph.swe_revjul(jd,CalendarType.SE_GREG_CAL).microsecondsSinceEpoch);}
  for(final date in [[2016,12,31,23,59,60],[2017,1,1,0,0,0],[2024,2,29,0,0,0],[2023,2,29,0,0,0]]) {
    emit({'kind':'utc','input':date},()=>Sweph.swe_utc_to_jd(date[0],date[1],date[2],date[3],date[4],date[5].toDouble(),CalendarType.SE_GREG_CAL));
  }
  for(final solar in [true,false]) for(final back in [true,false]) {
    emit({'kind':'eclipse','solar':solar,'back':back,'jd':2460310.5},(){
      final e=solar?Sweph.swe_sol_eclipse_when_glob(2460310.5,SwephFlag.SEFLG_SWIEPH,EclipseFlag.SE_ECL_ALLTYPES_SOLAR,back):Sweph.swe_lun_eclipse_when(2460310.5,SwephFlag.SEFLG_SWIEPH,EclipseFlag.SE_ECL_ALLTYPES_LUNAR,back);
      return {'times':e.times,'type':e.eclipseType?.value};
    });
  }
  emit({'kind':'et','jd':2451545.0,'body':1,'flags':258},()=>coords(Sweph.swe_calc(2451545.0,HeavenlyBody.SE_MOON,SwephFlag(258))));
  Sweph.swe_close();
}
