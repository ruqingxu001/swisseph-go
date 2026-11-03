// SPDX-License-Identifier: AGPL-3.0-only
// Execute using the user's Flutter application package configuration.
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:flutter_test/flutter_test.dart';
import 'package:sweph/sweph.dart';
import 'package:astrology_app/core/swephx/calc_tools.dart';
import 'package:astrology_app/core/swephx/models/heavenly_body.dart' as app;
import 'package:astrology_app/core/swephx/models/ecliptic_system.dart';
class Files implements AssetLoader {
 Files(this.path);final String path;
 Future<Uint8List> load(String name)=>File('$path/$name').readAsBytes();
}
void main(){
 test('Export actual CalcTools outputs for independent Go comparison',() async {
  final path=Platform.environment['SWISSGO_TEST_EPHE_PATH']!;
  await Sweph.init(modulePath:Platform.environment['SWISSGO_REFERENCE_LIBRARY']!,epheFilesPath:path,epheAssets:['sepl_18.se1'],assetLoader:Files(path));
  Sweph.swe_set_sid_mode(SiderealMode.SE_SIDM_LAHIRI,SiderealModeFlag.SE_SIDBIT_NONE,0,0);
  final rows=<Map<String,dynamic>>[];
  final bodies={app.HeavenlyBody.sun:0,app.HeavenlyBody.moon:1,app.HeavenlyBody.mercury:2,app.HeavenlyBody.venus:3,app.HeavenlyBody.mars:4,app.HeavenlyBody.jupiter:5,app.HeavenlyBody.saturn:6,app.HeavenlyBody.uranus:7,app.HeavenlyBody.neptune:8,app.HeavenlyBody.pluto:9,app.HeavenlyBody.earth:14,app.HeavenlyBody.northNode:10,app.HeavenlyBody.lilith:12,app.HeavenlyBody.chiron:15,app.HeavenlyBody.pholus:16,app.HeavenlyBody.ceres:17,app.HeavenlyBody.pallas:18,app.HeavenlyBody.juno:19,app.HeavenlyBody.vesta:20,app.HeavenlyBody.psyche:10016,app.HeavenlyBody.eros:10433,app.HeavenlyBody.cupido:40,app.HeavenlyBody.hades:41,app.HeavenlyBody.zeus:42,app.HeavenlyBody.kronos:43,app.HeavenlyBody.apollon:44,app.HeavenlyBody.admetos:45,app.HeavenlyBody.vulkanus:46,app.HeavenlyBody.poseidon:47};
  for(final date in [DateTime.utc(1900,1,1),DateTime.utc(1989,7,1,3),DateTime.utc(2000,1,1,12),DateTime.utc(2024,2,29,23,59,59,999,999),DateTime.utc(2026,11,2,16)]) {
   final jd=CalcTools.calcJulianDayFromUtc(date);
   rows.add({'kind':'appJulday','instant':date.toIso8601String(),'value':jd});
   rows.add({'kind':'revjul','jd':jd,'value':CalcTools.calcJDToDatetime(jd).microsecondsSinceEpoch});
   for(final ecliptic in EclipticSystem.values){
    final flags=258|(ecliptic==EclipticSystem.sidereal?65536:0);
    for(final pair in bodies.entries){
     final pos=CalcTools.calcPlanetPosition(jd,pair.key,ecliptic);
     rows.add({'kind':'calc','jd':jd,'body':pair.value,'flags':flags,'value':[pos.longitude,pos.latitude,pos.distance,pos.speedInLongitude,pos.speedInLatitude,pos.speedInDistance]});
    }
    for(final sys in Hsys.values){
     final h=CalcTools.calcMyHouseData(jd,121.4737,31.2304,sys,ecliptic);
     rows.add({'kind':'appHouse','jd':jd,'system':sys.value,'systemName':sys.name,'flags':2|(flags&65536),'lat':31.2304,'lon':121.4737,'value':{'cusps':h.cusps,'angles':[h.ascendant,h.midheaven,h.descendant,h.immumCoeli,h.vertex,h.eastPoint]}});
    }
   }
  }
  expect(()=>CalcTools.calcJulianDayFromUtc(DateTime(2000,1,1)),throwsArgumentError);
  File(Platform.environment['SWISSGO_APP_REFERENCE']!).writeAsStringSync(rows.map(jsonEncode).join('\n')+'\n');
  Sweph.swe_close();
 });
}
