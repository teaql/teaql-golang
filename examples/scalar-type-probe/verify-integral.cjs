#!/usr/bin/env node
'use strict';
// Installed verbatim as examples/scalar-type-probe/verify-integral.cjs.
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {createOwnedRunner,ownedBytes}=require('../fix-e-alignment/owned-command.cjs');
const example=__dirname,repo=path.resolve(example,'../..'),sha=b=>crypto.createHash('sha256').update(b).digest('hex');
const requested=process.env.TEAQL_EXAMPLES_EVIDENCE_DIR;
const evidence=requested?path.join(path.resolve(requested),'scalar-integral-admission'):fs.mkdtempSync(path.join(os.tmpdir(),'teaql-go-integral-evidence-'));
if(requested){assert(!fs.existsSync(evidence),'never overwrite evidence');fs.mkdirSync(evidence,{recursive:true});}
const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'teaql-go-integral-build-')),runner=createOwnedRunner();
const report={schema:'teaql.go-integral-example/v2',modeledInteger32:true,passed:false,localRuntime:repo,commands:[],rounds:[],
  publicPublication:false,sharedCachesDeleted:false,sourceDeleted:false,testDatabasesDeleted:false};
const cases=['integer-positive-overflow','integer-negative-overflow','optional-integer-positive-overflow','optional-integer-negative-overflow',
  'fractional-signed','signed-overflow','signed-float','signed-text','signed-bool','fractional-reference',
  'negative-reference','reference-overflow','reference-float','reference-text','reference-bool'];
async function main(){
try{
  const disk=fs.statfsSync(scratch);assert(disk.bavail*disk.bsize>4*1024**3);
  const provenance=fs.readFileSync(path.join(example,'generated-provenance.json')),generated=JSON.parse(provenance);
  assert.equal(generated.evaluationErrors,0);assert.equal(sha(fs.readFileSync(path.join(example,'model.xml'))),generated.modelSha256);
  for(const [file,hash]of Object.entries(generated.files))assert.equal(sha(fs.readFileSync(path.join(example,'generated',file))),hash);
  report.provenanceSha256=sha(provenance);fs.copyFileSync(path.join(example,'generated-provenance.json'),path.join(evidence,'generated-provenance.json'));
  const env={...process.env,GOWORK:'off',GOCACHE:path.join(scratch,'go-build'),GOTMPDIR:path.join(scratch,'go-tmp'),GOMAXPROCS:'2',TEAQL_FULL_SCALAR_DB:path.join(evidence,'scalars.sqlite')};fs.mkdirSync(env.GOTMPDIR);
  const manifest=fs.readFileSync(path.join(example,'go.mod'),'utf8')+'\nreplace scalar-type-conformance-service-core-workspace/lib => '+JSON.stringify(path.join(example,'generated'))+'\nreplace github.com/teaql/teaql-golang => '+JSON.stringify(repo)+'\n';
  fs.writeFileSync(path.join(scratch,'go.mod'),manifest);fs.copyFileSync(path.join(repo,'go.sum'),path.join(scratch,'go.sum'));
  report.applicationSources={};
  for(const [source,destination]of [['scalars_test.go','scalars_test.go'],['testdata/integral_admission_test.go','integral_admission_test.go']]){
    const bytes=fs.readFileSync(path.join(example,source));report.applicationSources[destination]=sha(bytes);fs.writeFileSync(path.join(scratch,destination),bytes);
  }
  async function run(name,command){
    const file=path.join(evidence,name+'.log'),fd=fs.openSync(file,'wx');let result;
    try{result=await runner.run(command,{cwd:scratch,env,timeout:600000,stdio:['ignore',fd,fd]});}finally{fs.closeSync(fd);}
    const log=fs.readFileSync(file,'utf8');report.commands.push({name,command,exitCode:result.status,signal:result.signal,logSha256:sha(Buffer.from(log))});
    assert.equal(result.status,0,log.slice(-8000));assert.equal(result.signal,null);assert(runner.safeToClean);return log;
  }
  const compiler=process.env.GO||'go',binary=path.join(scratch,'verify.test');
  await run('build',[compiler,'test','-race','-mod=mod','-c','-o',binary,'.']);report.binarySha256=sha(fs.readFileSync(binary));
  assert.equal((await run('runtime-origin',[compiler,'list','-m','-mod=mod','-f','{{.Dir}}','github.com/teaql/teaql-golang'])).trim(),repo);
  fs.copyFileSync(path.join(scratch,'go.mod'),path.join(evidence,'consumer-go.mod'));fs.copyFileSync(path.join(scratch,'go.sum'),path.join(evidence,'consumer-go.sum'));
  for(const round of [1,2]){
    const log=await run('round-'+round,[binary,'-test.v','-test.count=1','-test.run=^TestGeneratedIntegralAdmissionAfterCustomFix$']);
    assert(!/^--- FAIL:|panic:|^FAIL(?:\s|$)|DATA RACE/m.test(log));assert(/^PASS$/m.test(log));
    const rows=[...log.matchAll(/^CORE_INTEGRAL_REJECT language=go operation=(create|update) case=([a-z-]+) field=(required_integer|optional_integer|required_long|type_platform_id) query=0 mutate=0 begin=0 next_id=0 fix_calls=1$/gm)];
    assert.deepEqual(rows.map(m=>m[1]+':'+m[2]).sort(),['create','update'].flatMap(op=>cases.map(name=>op+':'+name)).sort());
    for(const row of rows)assert(new RegExp('^\\s+--- PASS: TestGeneratedIntegralAdmissionAfterCustomFix/'+row[1]+'/'+row[2]+' \\(','m').test(log));
    assert(log.includes('CORE_INTEGRAL_ORACLE rows='+round+' version=1 exact_long=true known_fk=true calibrated=true'));
    report.rounds.push({round,rejectedCases:30,rows:round});process.stdout.write(log);
  }
  const database='scalars.sqlite.integral.sqlite';report.database={file:database,sha256:sha(fs.readFileSync(path.join(evidence,database)))};
  assert.equal(sha(fs.readFileSync(path.join(example,'generated-provenance.json'))),report.provenanceSha256);
  report.passed=true;console.log('PASS: generated Go integral admission example twice without reset');
}catch(error){report.error=error.stack;console.error(error.stack);process.exitCode||=1;}
finally{
  report.disposedOwnedBytes=runner.safeToClean?ownedBytes(scratch):0;
  if(runner.safeToClean)fs.rmSync(scratch,{recursive:true});else report.unsafeBuildRetention={path:scratch,processGroup:runner.unsafeGroup};
  report.scratchRemoved=!fs.existsSync(scratch);runner.dispose();fs.writeFileSync(path.join(evidence,'receipt.json'),JSON.stringify(report,null,2)+'\n');
  console.log('Retained integral example evidence: '+evidence);
}}
main().catch(error=>{console.error(error.stack);process.exitCode||=1;});
