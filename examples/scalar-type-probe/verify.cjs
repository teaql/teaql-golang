#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),crypto=require('node:crypto'),assert=require('node:assert/strict');
const {createOwnedRunner,ownedBytes}=require('../fix-e-alignment/owned-command.cjs');
const example=__dirname,repo=path.resolve(example,'../..'),sha=bytes=>crypto.createHash('sha256').update(bytes).digest('hex');
const requested=process.env.TEAQL_EXAMPLES_EVIDENCE_DIR;
const evidence=requested?path.join(path.resolve(requested),'scalar-type-probe'):fs.mkdtempSync(path.join(os.tmpdir(),'teaql-go-scalar-evidence-'));
if(requested){assert(!fs.existsSync(evidence),'never replace a previous scalar run');fs.mkdirSync(evidence,{recursive:true});}
const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'teaql-go-scalar-build-')),runner=createOwnedRunner();
const report={schema:'teaql.go-scalar-example/v4',passed:false,localRuntime:repo,commands:[],rounds:[],typedEntityJsonAccepted:false,integerReadAccepted:false,booleanAccepted:false,
  publicPublication:false,sourceDeleted:false,sharedCachesDeleted:false,testDatabasesDeleted:false};
function fingerprint(){
  const bytes=fs.readFileSync(path.join(example,'generated-provenance.json')),m=JSON.parse(bytes);
  assert.equal(m.modelSha256,sha(fs.readFileSync(path.join(example,'model.xml'))));assert.equal(m.evaluationErrors,0);
  const expected=Object.keys(m.files).sort(),actual=[];
  function visit(dir,prefix=''){for(const e of fs.readdirSync(dir,{withFileTypes:true})){assert(!e.isSymbolicLink());
    const relative=path.join(prefix,e.name),file=path.join(dir,e.name);
    if(e.isDirectory())visit(file,relative);else{assert(e.isFile());actual.push(relative);}}}
  visit(path.join(example,'generated'));assert.deepEqual(actual.sort(),expected);
  for(const [file,hash]of Object.entries(m.files))assert.equal(sha(fs.readFileSync(path.join(example,'generated',file))),hash,file);
  return sha(bytes);
}
async function main(){
try{
  const disk=fs.statfsSync(scratch);assert(disk.bavail*disk.bsize>4*1024**3,'4 GiB free scratch required');
  report.provenanceSha256=fingerprint();
  fs.copyFileSync(path.join(example,'generated-provenance.json'),path.join(evidence,'generated-provenance.json'));
  const compiler=process.env.GO||'go';
  const env={...process.env,GOWORK:'off',GOCACHE:path.join(scratch,'go-build'),GOTMPDIR:path.join(scratch,'go-tmp'),
    GOMAXPROCS:'2',TEAQL_FULL_SCALAR_DB:path.join(evidence,'scalars.sqlite')};fs.mkdirSync(env.GOTMPDIR);
  const manifest=fs.readFileSync(path.join(example,'go.mod'),'utf8')+
    '\nreplace scalar-type-conformance-service-core-workspace/lib => '+JSON.stringify(path.join(example,'generated'))+
    '\nreplace github.com/teaql/teaql-golang => '+JSON.stringify(repo)+'\n';
  fs.writeFileSync(path.join(scratch,'go.mod'),manifest);report.applicationSources={};
  for(const file of fs.readdirSync(example).filter(file=>file.endsWith('_test.go'))){const bytes=fs.readFileSync(path.join(example,file));
    fs.copyFileSync(path.join(example,file),path.join(scratch,file));report.applicationSources[file]=sha(bytes);}
  async function run(name,command,timeout=300000){
    const file=path.join(evidence,name+'.log'),fd=fs.openSync(file,'wx');let result;
    try{result=await runner.run(command,{cwd:scratch,env,timeout,stdio:['ignore',fd,fd]});}finally{fs.closeSync(fd);}
    const log=fs.readFileSync(file,'utf8');report.commands.push({name,command,exitCode:result.status,signal:result.signal,logSha256:sha(Buffer.from(log))});
    assert.equal(result.status,0,log.slice(-12000));assert.equal(result.signal,null);assert(runner.safeToClean);return log;
  }
  const binary=path.join(scratch,'verify.test');
  await run('build',[compiler,'test','-mod=mod','-c','-o',binary,'.']);report.binarySha256=sha(fs.readFileSync(binary));
  const resolution=await run('runtime-origin',[compiler,'list','-m','-mod=mod','-f','{{.Dir}}','github.com/teaql/teaql-golang']);
  assert.equal(fs.realpathSync(resolution.trim()),fs.realpathSync(repo));
  fs.copyFileSync(path.join(scratch,'go.mod'),path.join(evidence,'consumer-go.mod'));
  if(fs.existsSync(path.join(scratch,'go.sum')))fs.copyFileSync(path.join(scratch,'go.sum'),path.join(evidence,'consumer-go.sum'));
  for(const round of [1,2]){
    const log=await run('round-'+round,[binary,'-test.v','-test.count=1'],90000);
    assert(!/^--- FAIL:|panic:|^FAIL(?:\s|$)/m.test(log));
    assert.equal((log.match(/^CORE_SCALAR_NATIVE .*timeProfile=TimeOfDay /gm)||[]).length,1);
    const cases=[...log.matchAll(/^CORE_NONFINITE_REJECT language=go field=(required_float|optional_float|required_double|optional_double) case=([012]) query=0 mutate=0 begin=0 next_id=0$/gm)];
    assert.equal(cases.length,12);assert.equal(new Set(cases.map(x=>x[1]+':'+x[2])).size,12);
    const time=[...log.matchAll(/^GO_TYPED_TIME=exact_nanoseconds:true null_preserved:true unsaved_timestamp:true pure_no_io:true predicate:true corrupt_legacy_rejected:true id=([1-9][0-9]*)$/gm)];
    assert.equal(time.length,1);assert(/^--- PASS: TestGeneratedTypedTimeIntegrity /m.test(log));assert(/^PASS$/m.test(log));
    const json=[...log.matchAll(/^GO_ENTITY_JSON profile=(camelCase|snake_case|PascalCase) fields=21 exact=true null_not_loaded=true forward=true reverse=true independent_roots=true stale_rejected=true partial_zero_io=true json_zero_io=true id=([1-9][0-9]*)$/gm)];
    assert.equal(json.length,3);assert.deepEqual(json.map(m=>m[1]).sort(),['PascalCase','camelCase','snake_case']);
    assert.equal(new Set(json.map(m=>m[2])).size,3);assert(/^--- PASS: TestGeneratedTypedEntityJSON /m.test(log));
    assert.equal((log.match(/^CORE_INTEGER_READ_REJECT /gm)||[]).length,40);
    assert(log.includes('CORE_INTEGER_READ_ORACLE rows='+(round*9)+' version=1 corrupt='+(round*8)+' queries=8 hydration=8 expressions=24 raw_values_redacted=true'));
    assert(/^--- PASS: TestGeneratedIntegerReadBoundaries \(/m.test(log));
    const boolean=[...log.matchAll(/^GO_BOOLEAN_BOUNDARY=(.+)$/gm)].map(m=>JSON.parse(m[1]));assert.equal(boolean.length,1);const b=boolean[0];assert(b.passed&&b.rejectedFixLedgerUnchanged&&b.trueFalseUnknown&&b.notLoadedIsNotNull&&b.repairedBooleanPersisted&&b.legacyIntegerRejected);assert.equal(b.invalidFixCases,12);assert.equal(b.rejectedFixProviderEntries,0);assert.equal(b.checks.length,12);assert.equal(new Set(b.checks).size,12);assert.equal(b.version,3);
    report.rounds.push({round,timeEntityId:time[0][1],nonFiniteCases:12,jsonEntityIDs:Object.fromEntries(json.map(m=>[m[1],m[2]])),integerRead:{queryFailures:8,atomicHydrationFailures:8,expressionFailures:24,rows:round*9,corruptRows:round*8,rawValuesRedacted:true},boolean:b});assert.equal(fingerprint(),report.provenanceSha256);process.stdout.write(log);
  }
  assert.notEqual(report.rounds[0].timeEntityId,report.rounds[1].timeEntityId,'SQLite must not reset between rounds');
  assert.equal(new Set(report.rounds.flatMap(r=>Object.values(r.jsonEntityIDs))).size,6,'JSON database must not reset between rounds');
  assert.equal(new Set(report.rounds.flatMap(r=>[r.boolean.entityId,r.boolean.repairedId])).size,4,'Boolean database must not reset between rounds');
  report.databases=['scalars.sqlite','scalars.sqlite.nonfinite.sqlite','scalars.sqlite.typed-time.sqlite','scalars.sqlite.entity-json.sqlite','scalars.sqlite.integer-read.sqlite','scalars.sqlite.boolean.sqlite'].map(file=>({file,sha256:sha(fs.readFileSync(path.join(evidence,file)))}));
  report.typedEntityJsonAccepted=true;report.integerReadAccepted=true;report.booleanAccepted=true;report.passed=true;console.log('PASS: generated Go scalar/TimeOfDay/typed-entity-JSON/Integer-read/Boolean example twice without cleanup');
}catch(error){report.error=error.stack;console.error(error.stack);process.exitCode||=1;}
finally{
  report.disposedOwnedBytes=runner.safeToClean?ownedBytes(scratch):0;
  if(runner.safeToClean)fs.rmSync(scratch,{recursive:true});else report.unsafeBuildRetention={path:scratch,processGroup:runner.unsafeGroup};
  report.scratchRemoved=!fs.existsSync(scratch);runner.dispose();
  fs.writeFileSync(path.join(evidence,'receipt.json'),JSON.stringify(report,null,2)+'\n');console.log('Retained scalar example evidence: '+evidence);
}}
main().catch(error=>{console.error(error.stack);process.exitCode||=1;});
