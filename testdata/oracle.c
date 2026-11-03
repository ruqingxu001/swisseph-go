/* SPDX-License-Identifier: AGPL-3.0-only
 * Native C reference harness. Test-only; never linked into the Go library.
 */
#include "swephexp.h"
#include <stdio.h>
#include <string.h>
static void values(double *p,int n) {printf("[");for(int i=0;i<n;i++)printf("%s%.17g",i?",":"",p[i]);printf("]");}
static void setup(char *path,int sid) {swe_set_ephe_path(path);swe_set_sid_mode(sid,0,0);swe_set_topo(121.4737,31.2304,20);}
int main(int argc,char **argv) {
 if(argc!=2)return 2;
 char *path=argv[1],msg[256];
 const double dates[]={625000.5,1500000.5,2299160.5,2415020.5,2447892.5,2451545,2458849.5,2460409.25,2461346.5,2462502.5,2816787.5};
 const int flags[]={SEFLG_MOSEPH|SEFLG_SPEED,SEFLG_SWIEPH|SEFLG_SPEED,SEFLG_MOSEPH|SEFLG_SPEED|SEFLG_SIDEREAL,SEFLG_SWIEPH|SEFLG_SPEED|SEFLG_EQUATORIAL,SEFLG_MOSEPH|SEFLG_SPEED|SEFLG_TOPOCTR,SEFLG_MOSEPH|SEFLG_SPEED|SEFLG_XYZ|SEFLG_J2000};
 for(int d=0;d<11;d++)for(int p=-1;p<=20;p++)for(int f=0;f<6;f++){
  double x[6]={0};msg[0]=0;setup(path,SE_SIDM_LAHIRI);
  int ret=swe_calc_ut(dates[d],p,flags[f],x,msg);
  printf("{\"kind\":\"calc\",\"jd\":%.17g,\"body\":%d,\"flags\":%d,\"sid\":1,\"status\":%d,\"v\":",dates[d],p,flags[f],ret);values(x,6);puts("}");swe_close();
 }
 // ET exercises a different entry point with a successful file-based Moon.
 {double x[6];setup(path,0);int ret=swe_calc(2451545,1,258,x,msg);printf("{\"kind\":\"et\",\"jd\":2451545,\"body\":1,\"flags\":258,\"status\":%d,\"v\":",ret);values(x,6);puts("}");swe_close();}
 const char *systems="PKORCAEVXHTBMGUWFIiJLYSQDN";
 const double lats[]={0,31.2304,-33.8688,70,-70};
 for(int j=0;j<3;j++)for(int i=0;systems[i];i++)for(int k=0;k<5;k++)for(int s=0;s<2;s++) {
  double c[37]={0},a[10]={0},cs[37]={0},as[10]={0};msg[0]=0;setup(path,SE_SIDM_LAHIRI);
  int flag=s?SEFLG_SIDEREAL:0;int n=systems[i]=='G'?37:13;
  int ret=swe_houses_ex2(dates[4+j],flag,lats[k],121.4737,systems[i],c,a,cs,as,msg);
  printf("{\"kind\":\"houses\",\"jd\":%.17g,\"flags\":%d,\"sid\":1,\"lat\":%.17g,\"lon\":121.4737,\"system\":%d,\"status\":%d,\"v\":",dates[4+j],flag,lats[k],systems[i],ret);values(c,n);printf(",\"a\":");values(a,10);printf(",\"cs\":");values(cs,n);printf(",\"as\":");values(as,10);puts("}");swe_close();
 }
 const int years[]={-4000,-1,0,1582,1900,2000,2024,2100};
 for(int i=0;i<8;i++)for(int cal=0;cal<2;cal++){
  setup(path,0);double jd=swe_julday(years[i],3,1,12.345,cal),h;int y,m,d;
  /* separately print the reverse call's hour (all pointers required). */
  swe_revjul(jd,cal,&y,&m,&d,&h);
  printf("{\"kind\":\"date\",\"jd\":%.17g,\"calendar\":%d,\"year\":%d,\"v\":[%d,%d,%d,%.17g]}\n",jd,cal,years[i],y,m,d,h);swe_close();
 }
 const int utc[][6]={{2016,12,31,23,59,60},{2017,1,1,0,0,0},{1972,6,30,23,59,60},{1960,1,1,12,0,0},{2024,2,29,0,0,0},{2023,2,29,0,0,0},{2024,1,1,23,59,60}};
 for(int i=0;i<7;i++){
  setup(path,0);double v[2]={0};int ret=swe_utc_to_jd(utc[i][0],utc[i][1],utc[i][2],utc[i][3],utc[i][4],utc[i][5],1,v,msg);
  printf("{\"kind\":\"utc\",\"status\":%d,\"input\":[%d,%d,%d,%d,%d,%d],\"v\":",ret,utc[i][0],utc[i][1],utc[i][2],utc[i][3],utc[i][4],utc[i][5]);values(v,2);puts("}");swe_close();
 }
 for(int solar=0;solar<2;solar++)for(int back=0;back<2;back++) {
  setup(path,0);double v[10]={0};int ret=solar?swe_sol_eclipse_when_glob(2460310.5,SEFLG_MOSEPH,0,v,back,msg):swe_lun_eclipse_when(2460310.5,SEFLG_MOSEPH,0,v,back,msg);
  printf("{\"kind\":\"eclipse\",\"jd\":2460310.5,\"solar\":%d,\"back\":%d,\"status\":%d,\"v\":",solar,back,ret);values(v,10);puts("}");swe_close();
 }
 return 0;
}
