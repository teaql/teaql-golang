'use strict';
const assert=require('node:assert/strict'),{failure}=require('./expression-boundaries.cjs');
const NAMES=['early-repeated','middle-list','late-repeated','middle-size','loaded-repeated','loaded-multiword','loaded-fk',
  'loaded-empty-and-out-of-range','loaded-null','fallback-not-loaded','shared-relation-siblings','no-side-effects'];
function validateRecursiveExpressions(log){
 assert(!/--- FAIL:|--- SKIP:|DATA RACE|panic:/m.test(log));
 assert.equal((log.match(/^--- PASS: TestGeneratedRecursiveExpressionBoundaries /gm)||[]).length,1);
 const frames=key=>log.split(/\r?\n/).filter(line=>line.startsWith(key+'=')).map(line=>JSON.parse(line.slice(key.length+1)));
 const rows=frames('GO_RECURSIVE_E_CASE'),results=frames('GO_RECURSIVE_E_RESULT');assert.equal(results.length,1);
 const r=results[0];assert.equal(r.passed,true);assert.equal(r.generatedCases,12);assert.deepEqual(r.cases,NAMES);
 assert(Number.isSafeInteger(r.parentId)&&r.parentId>0);assert.deepEqual(rows.map(v=>v.name),NAMES);assert(rows.every(v=>v.passed));
 const v=Object.fromEntries(rows.map(row=>[row.name,row.value])),root='E.PurchaseOrder(<root>)';
 const repeated=['tenant','purchase_order_list',0,'tenant','name'];
 failure(v['early-repeated'],['tenant'],repeated,'/tenant',0,
  root+'.Tenant()<broken>.PurchaseOrderList().First().Tenant().Name().TryEval()','Tenant','/tenant/purchaseOrderList/0/tenant/name');
 failure(v['middle-list'],['tenant','purchase_order_list'],['tenant','purchase_order_list',2,'tenant','name'],'/tenant/purchaseOrderList',1,
  root+'.Tenant().PurchaseOrderList()<broken>.Get(2).Tenant().Name().TryEval()','Tenant.PurchaseOrderList','/tenant/purchaseOrderList/2/tenant/name');
 failure(v['late-repeated'],repeated,repeated,'/tenant/purchaseOrderList/0/tenant/name',4,
  root+'.Tenant().PurchaseOrderList().First().Tenant().Name()<broken>.TryEval()','Tenant.PurchaseOrderList[0].Tenant.Name');
 failure(v['middle-size'],['tenant','purchase_order_list'],['tenant','purchase_order_list'],'/tenant/purchaseOrderList',1,
  root+'.Tenant().PurchaseOrderList()<broken>.Size().TryEval()','Tenant.PurchaseOrderList');
 assert.deepEqual(v['loaded-repeated'],{name:'Tenant Name',present:true});assert.deepEqual(v['loaded-multiword'],{value:'5550971',present:true});
 assert.deepEqual(v['loaded-fk'],{value:1,present:true});assert.deepEqual(v['loaded-empty-and-out-of-range'],{size:0,present:false});
 assert.deepEqual(v['loaded-null'],{value:null});
 failure(v['fallback-not-loaded'],['tenant'],repeated,'/tenant',0,
  root+'.Tenant()<broken>.PurchaseOrderList().First().Tenant().Name().OrElse(<fallback>)','Tenant','/tenant/purchaseOrderList/0/tenant/name');
 const siblings=v['shared-relation-siblings'];assert.deepEqual(siblings.leftBefore,siblings.leftAfter);assert.deepEqual(siblings.leftBefore,v['early-repeated']);
 failure(siblings.right,['tenant'],['tenant','purchase_order_list',2,'contact_phone'],'/tenant',0,
  root+'.Tenant()<broken>.PurchaseOrderList().Get(2).ContactPhone().TryEval()','Tenant','/tenant/purchaseOrderList/2/contactPhone');
 const pure=v['no-side-effects'];assert.deepEqual(pure.effects,{statements:0,transactions:0,policies:0,audits:0});
 assert.deepEqual(Object.keys(pure.calibration).sort(),Object.keys(pure.effects).sort());assert(Object.values(pure.calibration).every(n=>Number.isSafeInteger(n)&&n>0));
 assert.equal(pure.geometryUnchanged,true);assert.equal(pure.pendingUnchanged,true);assert.equal(pure.ledgerIdentitiesUnchanged,true);
 assert(!JSON.stringify(rows).includes('PRIVATE_GO_'));return r;
}
module.exports={NAMES,validateRecursiveExpressions};
