'use strict';
const fs = require('node:fs'), path = require('node:path'), os = require('node:os');
const crypto = require('node:crypto'), assert = require('node:assert/strict'), cp = require('node:child_process');
const {createOwnedRunner, ownedBytes} = require('../fix-e-alignment/owned-command.cjs');
const repo = path.resolve(__dirname, '../..'), fixture = path.join(__dirname, 'fixture');
const sha = value => crypto.createHash('sha256').update(value).digest('hex');
const requested = process.env.TEAQL_CONTEXT_FIX_EVIDENCE_DIR;
const output = requested ? path.resolve(requested) : fs.mkdtempSync(path.join(os.tmpdir(), 'teaql-go-context-evidence-'));
if (requested) { assert(!fs.existsSync(output), 'use a new evidence directory'); fs.mkdirSync(output, {recursive: true}); }
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'teaql-go-context-build-')), runner = createOwnedRunner();
const report = {schema: 'teaql.go-context-derived-mutation/v1', passed: false, commands: [], rounds: [], publicPublication: false};
function inputs() {
  const bytes = fs.readFileSync(path.join(fixture, 'generated-files.json')), files = JSON.parse(bytes);
  for (const [name, hash] of Object.entries(files)) {
    assert(!path.isAbsolute(name) && !name.split('/').includes('..'));
    const file = path.join(fixture, name); assert.equal(fs.realpathSync(file), file);
    assert.equal(sha(fs.readFileSync(file)), hash, name);
  }
  return {files, manifestSha256: sha(bytes), modelSha256: sha(fs.readFileSync(path.join(fixture, 'model.xml')))};
}
function runtimeHash() {
  const hash = crypto.createHash('sha256');
  const files = cp.execFileSync('git', ['-C', repo, 'ls-files', '-z', '--cached', '--others', '--exclude-standard'])
    .toString().split('\0').filter(name => name && !name.startsWith('examples/')).sort();
  for (const name of [...new Set(files)]) {
    const file = path.join(repo,name); if (!fs.existsSync(file)) continue;
    assert(fs.lstatSync(file).isFile()); hash.update(name+'\0'); hash.update(fs.readFileSync(file));
  }
  return hash.digest('hex');
}
async function main() {
  try {
    const disk = fs.statfsSync(scratch); assert(disk.bavail * disk.bsize > 2 * 1024 ** 3);
    report.fixture = inputs(); report.sourceSha256 = runtimeHash();
    report.sourceCommit = cp.execFileSync('git', ['-C',repo,'rev-parse','HEAD']).toString().trim();
    const app = fs.readFileSync(path.join(__dirname,'context_fix_test.go')); report.applicationSha256 = sha(app);
    fs.writeFileSync(path.join(scratch,'context_fix_test.go'),app);
    const mod = fs.readFileSync(path.join(__dirname,'go.mod'),'utf8')
      .replace('=> ./fixture','=> '+JSON.stringify(fixture)).replace('=> ../..','=> '+JSON.stringify(repo));
    fs.writeFileSync(path.join(scratch,'go.mod'),mod);
    const env = {...process.env, GOWORK:'off', GOENV:'off', GOMAXPROCS:'2', GOTMPDIR:path.join(scratch,'tmp'),
      TEAQL_TELEMETRY_ENABLED:'false', TEAQL_CONTEXT_FIX_DB:path.join(output,'fixture.sqlite')};
    fs.mkdirSync(env.GOTMPDIR);
    for (const key of Object.keys(env)) if (/TOKEN|PASSWORD|SECRET|AUTH|API_KEY/.test(key)) delete env[key];
    const go = process.env.GO || 'go';
    async function run(name, command, timeout=300000) {
      const log = path.join(output,name+'.log'), fd = fs.openSync(log,'wx'); let result;
      try { result = await runner.run(command,{cwd:scratch,env,timeout,stdio:['ignore',fd,fd]}); }
      finally { fs.closeSync(fd); }
      const bytes = fs.readFileSync(log);
      report.commands.push({name,command,exitCode:result.status,signal:result.signal,logSha256:sha(bytes)});
      assert(!result.error && !result.signal && result.status===0,bytes.toString().slice(-6000));
      return bytes.toString();
    }
    report.origin = JSON.parse(await run('origin',[go,'list','-m','-json','github.com/teaql/teaql-golang']));
    assert.equal(report.origin.Replace?.Dir,repo);
    const binary=path.join(scratch,'context-fix.test');
    await run('build',[go,'test','-race','-mod=mod','-c','-o',binary,'.']);
    report.binarySha256=sha(fs.readFileSync(binary));
    for (const round of [1,2]) {
      const log=await run('round-'+round,[binary,'-test.v','-test.count=1','-test.timeout=60s'],90000);
      assert.match(log,/GO_CONTEXT_FIX_SQLITE=passed typedQ=true typedE=true noOpProviderCalls=0 failedCheckProviderCalls=0 updateAudits=1 objectLocation=true graphRollback=true/);
      assert(!log.includes('WARNING: DATA RACE'));
      assert.deepEqual(inputs(),report.fixture); assert.equal(runtimeHash(),report.sourceSha256);
      report.rounds.push({round,passed:true});
    }
    fs.copyFileSync(path.join(scratch,'go.mod'),path.join(output,'resolved-go.mod'));
    fs.copyFileSync(path.join(scratch,'go.sum'),path.join(output,'resolved-go.sum'));
    report.sameDatabase=true; report.databaseSha256=sha(fs.readFileSync(path.join(output,'fixture.sqlite')));
    report.passed=true;
  } catch(error) { report.error=error.stack; process.exitCode=1; console.error(error.stack); }
  finally {
    report.disposedOwnedBytes=runner.safeToClean?ownedBytes(scratch):0;
    if(runner.safeToClean) fs.rmSync(scratch,{recursive:true}); else report.unsafeBuildRetention=scratch;
    report.scratchRemoved=!fs.existsSync(scratch); runner.dispose();
    fs.writeFileSync(path.join(output,'receipt.json'),JSON.stringify(report,null,2)+'\n');
    console.log(JSON.stringify({passed:report.passed,rounds:report.rounds.length,evidence:output}));
    if(!requested && report.passed && report.scratchRemoved) fs.rmSync(output,{recursive:true});
  }
}
main().catch(error=>{console.error(error.stack);process.exitCode=1;});
