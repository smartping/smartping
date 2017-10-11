import xlsxwriter
import sqlite3
import socket
import sys
import os
import platform

SQLITFILE='database.db'

def getos():
    return platform.system()

def createdir():
    if getos()=="Windows" :
        outpath=sys.path[0]+'\out'
    else:
        outpath=sys.path[0]+'/out'
    if os.path.exists(outpath):
        print 'out dir is exists'
        return None
    else:
        os.makedirs(outpath)
        return outpath
    
def getpingtable():
    conn = sqlite3.connect(SQLITFILE)
    cursor = conn.cursor()
    tables=[]
    for row in cursor.execute('SELECT name FROM sqlite_master WHERE type=\'table\' and name like \'pinglog-%\' ORDER BY name;'):
        tables.append(row)
    conn.close()
    return tables

def gethostname():
    return socket.gethostname()

def dict_factory(cursor, row):
    dictlist = {}
    for idx, col in enumerate(cursor.description):
        dictlist[col[0]] = row[idx]
    return dictlist

def wirtexls(hostname,xlsfile):
    conn = sqlite3.connect(SQLITFILE)
    conn.row_factory = dict_factory
    cursor = conn.cursor()
    cursor.execute('SELECT maxdelay,mindelay,avgdelay,sendpk,revcpk,losspk,lastcheck FROM [pinglog-%s];'% xlsfile)
    if getos()=="Windows" :
        fullxlsfilename='out\\'+hostname+'-to-'+xlsfile+'.xlsx'
    else:
        fullxlsfilename='out//'+hostname+'-to-'+xlsfile+'.xlsx'  
    workbook = xlsxwriter.Workbook(fullxlsfilename)
    bold = workbook.add_format({'bold': True})
    worksheet = workbook.add_worksheet()
    worksheet.write('A1', 'from',bold)
    worksheet.write('B1', 'maxdelay',bold)
    worksheet.write('C1', 'mindelay',bold)
    worksheet.write('D1', 'avgdelay',bold)
    worksheet.write('E1', 'sendpk',bold)
    worksheet.write('F1', 'revcpk',bold)
    worksheet.write('G1', 'losspk',bold)
    worksheet.write('H1', 'lastcheck',bold)
    xlsrow=1;
    xlscol=0;
    for row in cursor.fetchall():
        worksheet.write(xlsrow,xlscol,hostname)
        worksheet.write(xlsrow,xlscol+1,row['maxdelay'])
        worksheet.write(xlsrow,xlscol+2,row['mindelay'])
        worksheet.write(xlsrow,xlscol+3,row['avgdelay'])
        worksheet.write(xlsrow,xlscol+4,row['sendpk'])
        worksheet.write(xlsrow,xlscol+5,row['revcpk'])
        worksheet.write(xlsrow,xlscol+6,row['losspk'])
        worksheet.write(xlsrow,xlscol+7,row['lastcheck'])
        xlsrow+=1
    workbook.close()
    conn.close()
    return True
    
if __name__=="__main__":
    createdir();
    for tablename in getpingtable():
        tablestr=''.join(tablename)
        xlsfile= tablestr.replace('pinglog-','')
        wirtexls(gethostname(),xlsfile)
    print "over!!!"
