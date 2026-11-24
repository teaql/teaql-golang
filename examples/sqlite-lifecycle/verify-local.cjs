'use strict';

const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {createOwnedRunner, ownedBytes} = require('../fix-e-alignment/owned-command.cjs');

const example = __dirname;
const repo = path.resolve(example, '../..');
const parent = path.join(repo, 'examples/fix-e-alignment');
const workspace = path.join(example, 'golang-app-console');
const library = path.join(parent, 'golang-lib-core');
const requested = process.env.TEAQL_SQLITE_LIFECYCLE_EVIDENCE_DIR;
const cases = ['replacement', 'fanout', 'idle-eviction', 'transaction-cancellation', 'runtime-origin'];
const sha = data => crypto.createHash('sha256').update(data).digest('hex');
const shaFile = file => sha(fs.readFileSync(file));

function checkLibrary() {
  const manifestPath = path.join(parent, 'fixture-manifest.json');
  const manifest = JSON.parse(fs.readFileSync(manifestPath));
  const model = path.join(parent, 'model.xml');
  assert.equal(shaFile(model), manifest.modelSha256);
  for (const [name, hash] of Object.entries(manifest.files)) {
    const file = path.resolve(library, name);
    assert(file.startsWith(library + path.sep));
    assert.equal(shaFile(file), hash, name);
  }
  return {fixtureManifestSHA256: shaFile(manifestPath), modelSHA256: shaFile(model)};
}

function checkRound(log, providerHash) {
  assert(!log.includes('WARNING: DATA RACE'));
  const rows = log.split('\n').filter(line => line.startsWith('SQLITE_LIFECYCLE_CASE='))
    .map(line => JSON.parse(line.slice('SQLITE_LIFECYCLE_CASE='.length)));
  assert.deepEqual(rows.map(row => row.case), cases);
  for (const {facts} of rows.slice(0, 4)) {
    assert(facts.generatedSavePassed);
    assert(facts.initial.every(value => value.passed));
    assert(facts.observations.flat().every(value => value.passed));
    assert.equal(facts.callerContextReplaced, false);
  }
  assert.equal(rows[4].facts.sourceFallback, false);
  assert.equal(rows[4].facts.sha256, providerHash);
  return rows;
}

async function main() {
  let output, scratch, runner;
  const report = {schema: 'teaql.go-sqlite-lifecycle-example/v1', passed: false, commands: [], rounds: [],
    localRuntime: repo, sameDatabaseAcrossRounds: true, publicPublication: false};
  try {
    if (requested) {
      assert(path.isAbsolute(requested));
      fs.mkdirSync(requested);
      output = requested;
    } else {
      output = fs.mkdtempSync(path.join(os.tmpdir(), 'teaql-sqlite-example-evidence-'));
    }
    scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'teaql-sqlite-example-build-'));
    runner = createOwnedRunner();
    const disk = fs.statfsSync(scratch);
    assert(disk.bavail * disk.bsize > 2 * 1024 ** 3);
    const provider = path.join(repo, 'provider/sqlite/sqlite.go');
    const probe = path.join(workspace, 'lifecycle_test.go');
    const providerHash = shaFile(provider);
    report.inputs = {...checkLibrary(), providerSha256: providerHash, applicationSha256: shaFile(probe)};
    const consumer = path.join(scratch, 'workspace');
    fs.cpSync(workspace, consumer, {recursive: true});
    const modulePath = path.join(consumer, 'go.mod');
    const originalModule = fs.readFileSync(modulePath, 'utf8');
    const moduleText = originalModule
      .replace('=> ../../fix-e-alignment/golang-lib-core\n', '=> ' + JSON.stringify(library) + '\n')
      .replace('=> ../../..\n', '=> ' + JSON.stringify(repo) + '\n');
    assert(!moduleText.includes('=> ../'));
    fs.writeFileSync(modulePath, moduleText);
    fs.writeFileSync(path.join(output, 'consumer-go.mod'), moduleText);

    const env = {...process.env, GOWORK: 'off', GOENV: 'off', GOCACHE: path.join(scratch, 'cache'),
      GOTMPDIR: path.join(scratch, 'tmp'), GOMAXPROCS: '2', TEAQL_TELEMETRY_ENABLED: 'false',
      CORE236_SQLITE_RUNTIME_ROOT: repo, CORE236_SQLITE_PROVIDER_SHA256: providerHash,
      CORE236_SQLITE_LIFECYCLE_DB: path.join(output, 'lifecycle')};
    for (const name of Object.keys(env)) if (/TOKEN|PASSWORD|SECRET|AUTH|API_KEY/.test(name)) delete env[name];
    fs.mkdirSync(env.GOTMPDIR);
    const go = process.env.GO || 'go';
    async function run(name, command, timeout = 300000) {
      const logfile = path.join(output, name + '.log');
      const fd = fs.openSync(logfile, 'wx');
      let result;
      try {
        result = await runner.run(command, {cwd: consumer, env, timeout, stdio: ['ignore', fd, fd]});
      } finally {
        fs.closeSync(fd);
      }
      const log = fs.readFileSync(logfile, 'utf8');
      report.commands.push({name, command, exitCode: result.status, signal: result.signal, logSha256: shaFile(logfile)});
      assert(!result.error && !result.signal && result.status === 0, name + ': ' + log.slice(-4000));
      return log;
    }
    report.origin = JSON.parse(await run('origin', [go, 'list', '-m', '-json', 'github.com/teaql/teaql-golang']));
    assert.equal(report.origin.Replace?.Dir, repo);
    report.domainOrigin = JSON.parse(await run('domain-origin', [go, 'list', '-m', '-json', 'checker-fix-advanced-core-workspace/lib']));
    assert.equal(report.domainOrigin.Replace?.Dir, library);
    await run('build-console', [go, 'build', '-race', '-mod=mod', '-o', path.join(scratch, 'example'), './src']);
    const binary = path.join(scratch, 'lifecycle.test');
    await run('build', [go, 'test', '-race', '-mod=mod', '-c', '-o', binary, '.']);
    report.binarySha256 = shaFile(binary);
    fs.copyFileSync(modulePath, path.join(output, 'resolved-go.mod'));
    fs.copyFileSync(path.join(consumer, 'go.sum'), path.join(output, 'consumer-go.sum'));
    for (const round of [1, 2]) {
      const log = await run('round-' + round, [binary, '-test.v', '-test.count=1', '-test.timeout=60s',
        '-test.run', '^TestCore236SQLite'], 90000);
      report.rounds.push({round, observations: checkRound(log, providerHash)});
    }
    report.stores = {};
    for (const file of fs.readdirSync(output).filter(name => name.endsWith('.sqlite'))) {
      report.stores[file] = shaFile(path.join(output, file));
    }
    assert.equal(Object.keys(report.stores).length, 4);
    assert.equal(shaFile(provider), providerHash);
    assert.equal(shaFile(probe), report.inputs.applicationSha256);
    assert.deepEqual(checkLibrary(), {fixtureManifestSHA256: report.inputs.fixtureManifestSHA256,
      modelSHA256: report.inputs.modelSHA256});
    report.passed = true;
  } catch (error) {
    report.error = error.stack;
    process.exitCode = 1;
  } finally {
    report.cleanup = {scratchRemoved: false, sharedCachesDeleted: false, sourceDeleted: false};
    try {
      if (!runner || runner.safeToClean) {
        report.cleanup.disposedOwnedBytes = scratch ? ownedBytes(scratch) : 0;
        if (scratch) fs.rmSync(scratch, {recursive: true});
        report.cleanup.scratchRemoved = true;
      } else {
        report.cleanup.unsafeRetention = scratch;
        report.passed = false;
        process.exitCode = 1;
      }
    } catch (error) {
      report.cleanup.error = error.stack;
      report.passed = false;
      process.exitCode = 1;
    } finally {
      runner?.dispose();
    }
    if (output) fs.writeFileSync(path.join(output, 'report.json'), JSON.stringify(report, null, 2) + '\n');
    console.log(JSON.stringify({passed: report.passed, rounds: report.rounds.length,
      evidence: requested && output, temporaryEvidenceRemoved: !requested && report.cleanup.scratchRemoved,
      error: report.error}));
    if (output && !requested && report.cleanup.scratchRemoved) fs.rmSync(output, {recursive: true});
  }
}

main().catch(error => {
  console.error(error.stack);
  process.exitCode = 1;
});
