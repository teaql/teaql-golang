'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),cp=require('node:child_process'),crypto=require('node:crypto'),assert=require('node:assert/strict');
const {createOwnedRunner,ownedBytes}=require('../examples/fix-e-alignment/owned-command.cjs');
const repo=path.resolve(__dirname,'..'),args=process.argv.slice(2),index=args.indexOf('--output');
assert(index>=0&&args[index+1],'--output /new/evidence/path is required');
const output=path.resolve(args[index+1]);assert(!fs.existsSync(output),'never overwrite retained evidence');fs.mkdirSync(output,{recursive:true});
const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'teaql-go-natural-build-')),runner=createOwnedRunner();
const sha=bytes=>crypto.createHash('sha256').update(bytes).digest('hex');
function source(){
  const files=[...new Set(cp.execFileSync('git',['ls-files','-z','--cached','--others','--exclude-standard'],{cwd:repo}).toString().split('\0').filter(Boolean))].sort();
  return Object.fromEntries(files.map(name=>{const file=path.join(repo,name);assert(fs.lstatSync(file).isFile()&&!fs.lstatSync(file).isSymbolicLink());return [name,sha(fs.readFileSync(file))];}));
}
const report={scope:'local-source',passed:false,commands:[],publicPublication:false};
const go=process.env.GO||'go',env={...process.env,GOWORK:'off',GOENV:'off',GOMAXPROCS:'2',GOTMPDIR:path.join(scratch,'tmp'),TEAQL_TELEMETRY_ENABLED:'false'};
fs.mkdirSync(env.GOTMPDIR);
for(const key of Object.keys(env))if(/TOKEN|PASSWORD|SECRET|AUTH|API_KEY/.test(key))delete env[key];
async function run(name,command,cwd=repo,timeout=300000,extra={}){
  const file=path.join(output,name+'.log'),fd=fs.openSync(file,'wx');let result;
  try{result=await runner.run(command,{cwd,env:{...env,...extra},timeout,stdio:['ignore',fd,fd]});}finally{fs.closeSync(fd);}
  const log=fs.readFileSync(file);report.commands.push({name,command,exitCode:result.status,signal:result.signal,log:path.basename(file),logSha256:sha(log)});
  assert(!result.error&&!result.signal&&result.status===0,`${name} failed; see ${file}`);return log.toString();
}
function testSummary(log){
  const events=log.trim().split('\n').filter(Boolean).map(line=>JSON.parse(line));
  return {passed:events.filter(e=>e.Action==='pass'&&e.Test).length,failed:events.filter(e=>e.Action==='fail'&&e.Test).length,
    skipped:events.filter(e=>e.Action==='skip'&&e.Test).length,packages:events.filter(e=>e.Action==='pass'&&!e.Test).length,
    packageNames:events.filter(e=>['pass','skip'].includes(e.Action)&&!e.Test).map(e=>e.Package).sort()};
}
function generated(){
  const fixture=path.join(repo,'examples/natural-messages/fixture'),bytes=fs.readFileSync(path.join(fixture,'generated-files.json')),files=JSON.parse(bytes);
  for(const [name,digest]of Object.entries(files)){assert(!path.isAbsolute(name)&&!name.split('/').includes('..'));const file=path.join(fixture,name);assert.equal(fs.realpathSync(file),file);assert.equal(sha(fs.readFileSync(file)),digest,name);}
  return {files,manifestSha256:sha(bytes),modelSha256:sha(fs.readFileSync(path.join(fixture,'model.xml')))};
}
async function main(){
  try{
    const disk=fs.statfsSync(scratch);assert(disk.bavail*disk.bsize>2*1024**3);report.freeBytesBefore=disk.bavail*disk.bsize;
    report.sourceCommit=cp.execFileSync('git',['rev-parse','HEAD'],{cwd:repo,encoding:'utf8'}).trim();report.sourceHashes=source();report.scratch=scratch;
    report.toolchain=(await run('toolchain',[go,'version'])).trim();
    report.paths=JSON.parse(await run('paths',[go,'env','-json','GOCACHE','GOMODCACHE','GOTMPDIR','GOENV','GOWORK']));
    if(!args.includes('--example-only')){
      const focused=await run('focused',[go,'test','-json','./runtime','-run','^(TestNaturalChecker|TestCheckerPrivacy|TestContextRootCatalog)','-count=1']);
      report.focused=testSummary(focused);assert(report.focused.passed>=24&&report.focused.failed===0);
      const golden=[...focused.matchAll(/GO_NATURAL_GOLDEN=([^\\\n]+)/g)].map(m=>m[1]);assert.equal(golden.length,24);
      if(args.includes('--full')){
        report.expectedTestPackages=(await run('test-packages',[go,'list','-f','{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}','./...'])).trim().split('\n').filter(Boolean).sort();
        report.native=testSummary(await run('native',[go,'test','-json','./...','-count=1']));assert(report.native.failed===0);
        for(const name of report.expectedTestPackages)assert(report.native.packageNames.includes(name),'native suite omitted '+name);
        report.race=testSummary(await run('race',[go,'test','-race','-json','./runtime','-count=1']));assert(report.race.failed===0&&report.race.passed>300);
      }
    }
    if(args.includes('--example')||args.includes('--example-only')){
      report.fixture=generated();const app=path.join(scratch,'consumer');fs.mkdirSync(app);
      fs.copyFileSync(path.join(repo,'examples/natural-messages/natural_messages_test.go'),path.join(app,'natural_messages_test.go'));
      const mod=fs.readFileSync(path.join(repo,'examples/natural-messages/go.mod'),'utf8')
        .replace('=> ./fixture','=> '+JSON.stringify(path.join(repo,'examples/natural-messages/fixture'))).replace('=> ../..','=> '+JSON.stringify(repo));
      fs.writeFileSync(path.join(app,'go.mod'),mod);
      report.runtimeOrigin=JSON.parse(await run('origin',[go,'list','-m','-json','github.com/teaql/teaql-golang'],app));assert.equal(report.runtimeOrigin.Replace?.Dir,repo);
      const binary=path.join(scratch,'natural.test');await run('example-build',[go,'test','-race','-mod=mod','-c','-o',binary,'.'],app);
      report.binarySha256=sha(fs.readFileSync(binary));report.rounds=[];
      for(const round of [1,2]){
        const log=await run('example-'+round,[binary,'-test.v','-test.count=1','-test.timeout=60s'],app,90000,{TEAQL_NATURAL_MESSAGES_DB:path.join(output,'example.sqlite')});
        const responses=[...log.matchAll(/^GO_NATURAL_RESPONSE=(.+)$/gm)].map(m=>JSON.parse(m[1]));assert.equal(responses.length,4);
        const frames=[...log.matchAll(/^GO_NATURAL_SQLITE_PASS=(.+)$/gm)].map(m=>JSON.parse(m[1]));assert.equal(frames.length,1);assert.equal(frames[0].passed,true);
        assert.equal(frames[0].providerCallsOnRejection,0);assert.equal(frames[0].transactionCallsOnRejection,0);assert.equal(frames[0].updatedAudits,1);
        assert(frames[0].automaticMetadata&&frames[0].originalGeneratedChecker&&frames[0].unchangedParent&&frames[0].requiredBeforeIds);
        assert.equal(frames[0].customerRendererCalls,1);
        fs.writeFileSync(path.join(output,'responses-'+round+'.json'),JSON.stringify(responses,null,2)+'\n');report.rounds.push({round,...frames[0]});
        assert.deepEqual(generated(),report.fixture);assert.deepEqual(source(),report.sourceHashes,'candidate changed during verification');
      }
      fs.copyFileSync(path.join(app,'go.mod'),path.join(output,'resolved-go.mod'));fs.copyFileSync(path.join(app,'go.sum'),path.join(output,'resolved-go.sum'));
      report.sameDatabase=true;report.databaseSha256=sha(fs.readFileSync(path.join(output,'example.sqlite')));
    }
    assert.deepEqual(source(),report.sourceHashes);report.passed=true;
  }catch(error){report.error=error.stack;process.exitCode=1;console.error(error.stack);}
  finally{
    report.removedBuildBytes=runner.safeToClean?ownedBytes(scratch):0;
    if(runner.safeToClean)fs.rmSync(scratch,{recursive:true});else report.unsafeBuildRetention=scratch;
    report.scratchRemoved=!fs.existsSync(scratch);runner.dispose();
    report.retainedFileHashes=Object.fromEntries(fs.readdirSync(output).filter(n=>fs.statSync(path.join(output,n)).isFile()).map(n=>[n,sha(fs.readFileSync(path.join(output,n)))]));
    fs.writeFileSync(path.join(output,'receipt.json'),JSON.stringify(report,null,2)+'\n');
    console.log(JSON.stringify({passed:report.passed,evidence:output,focused:report.focused,native:report.native,race:report.race,rounds:report.rounds?.length,removedBuildBytes:report.removedBuildBytes,scratchRemoved:report.scratchRemoved}));
  }
}
main().catch(error=>{console.error(error.stack);process.exitCode=1;});
