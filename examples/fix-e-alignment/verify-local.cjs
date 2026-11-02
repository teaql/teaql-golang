#!/usr/bin/env node
"use strict";
const fs = require("node:fs"), path = require("node:path"), os = require("node:os"), cp = require("node:child_process"), crypto = require("node:crypto"), assert = require("node:assert/strict");
const example = __dirname, repo = path.resolve(example, "../.."), sha = bytes => crypto.createHash("sha256").update(bytes).digest("hex");
const {validateExpressionBoundaries} = require('./expression-boundaries.cjs');
function fingerprint() {
  const m = JSON.parse(fs.readFileSync(path.join(example, 'fixture-manifest.json')));
  assert.equal(m.language, 'go'); assert.equal(m.evaluationErrors, 0);
  assert.equal(sha(fs.readFileSync(path.join(example, 'model.xml'))), m.modelSha256);
  const files = [];
  function visit(dir, prefix = '') { for (const e of fs.readdirSync(dir, {withFileTypes: true})) {
    const name = path.join(prefix, e.name), file = path.join(dir, e.name); assert(!e.isSymbolicLink());
    if (e.isDirectory()) visit(file, name); else { assert(e.isFile()); files.push(name); }
  } }
  visit(path.join(example, 'generated')); assert.deepEqual(files.sort(), Object.keys(m.files).sort());
  for (const [name, expected] of Object.entries(m.files)) assert.equal(sha(fs.readFileSync(path.join(example, 'generated', name))), expected, name);
  return m;
}
const requested = process.env.TEAQL_FIX_E_EVIDENCE_DIR;
const evidence = requested ? path.resolve(requested) : fs.mkdtempSync(path.join(os.tmpdir(), 'teaql-go-fix-e-evidence-'));
if (requested) { assert(path.isAbsolute(requested) && !fs.existsSync(evidence)); fs.mkdirSync(evidence, {recursive: true}); }
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'teaql-go-fix-e-build-'));
const report = {schema: 'teaql.go-fix-e-example/v1', passed: false, commands: [], sameDatabase: true,
  localRuntime: repo, databaseRetained: true, sharedCacheRemoved: false};
const {createOwnedRunner, ownedBytes} = require('./owned-command.cjs'), runner = createOwnedRunner();
async function main() {
try {
  const disk = fs.statfsSync(scratch); assert(disk.bavail * disk.bsize > 4 * 1024 ** 3);
  const m = fingerprint(); report.modelSha256 = m.modelSha256;
  report.contextRootApplicationSha256 = sha(fs.readFileSync(path.join(example, 'context_root_test.go')));
  report.partialAncestorApplicationSha256 = sha(fs.readFileSync(path.join(example, 'partial_ancestors_test.go')));
  report.partialRootApplicationSha256 = sha(fs.readFileSync(path.join(example, 'partial_root_test.go')));
  report.projectionApplicationSha256 = sha(fs.readFileSync(path.join(example, 'projection_views_test.go')));
  report.unchangedApplicationSha256 = sha(fs.readFileSync(path.join(example, 'unchanged_views_test.go')));
  report.rollbackApplicationSha256 = sha(fs.readFileSync(path.join(example, 'rollback_views_test.go')));
  report.checkerBoundaryApplicationSha256 = sha(fs.readFileSync(path.join(example, 'checker_boundaries_test.go')));
  report.expressionBoundaryApplicationSha256 = sha(fs.readFileSync(path.join(example, 'expression_boundary_test.go')));
  report.expressionBoundaryValidatorSha256 = sha(fs.readFileSync(path.join(example, 'expression-boundaries.cjs')));
  report.expressionBoundaryRounds = [];
  const manifest = fs.readFileSync(path.join(example, 'go.mod'), 'utf8')
    + '\nreplace checker-fix-advanced-core-workspace/lib => ' + JSON.stringify(path.join(example, 'generated'))
    + '\nreplace github.com/teaql/teaql-golang => ' + JSON.stringify(repo) + '\n';
  fs.writeFileSync(path.join(scratch, 'go.mod'), manifest); fs.writeFileSync(path.join(evidence, 'consumer-go.mod'), manifest);
  for (const file of fs.readdirSync(example).filter(file => file.endsWith('_test.go'))) fs.copyFileSync(path.join(example, file), path.join(scratch, file));
  const env = {...process.env, GOWORK: 'off', GOCACHE: path.join(scratch, 'cache'), GOTMPDIR: path.join(scratch, 'tmp'),
    GOMAXPROCS: '2', TEAQL_FIX_E_DB: path.join(evidence, 'fixture.sqlite')};
  fs.mkdirSync(env.GOTMPDIR);
  async function run(command, log, timeout) {
    const file = path.join(evidence, log), fd = fs.openSync(file, 'w'); let result;
    try { result = await runner.run(command, {cwd: scratch, env, timeout, stdio: ['ignore', fd, fd]}); }
    finally { fs.closeSync(fd); }
    const bytes = fs.readFileSync(file);
    report.commands.push({command, exitCode: result.status, signal: result.signal, logSha256: sha(bytes)});
    if (result.status !== 0) process.exitCode = result.status || 1;
    assert.equal(result.status, 0, bytes.toString().slice(-12000)); return bytes.toString();
  }
  const binary = path.join(scratch, 'verify.test');
  await run(['go', 'test', '-mod=mod', '-c', '-o', binary, '.'], 'build.log', 300000);
  report.binarySha256 = sha(fs.readFileSync(binary));
  fs.copyFileSync(path.join(scratch, 'go.mod'), path.join(evidence, 'resolved-go.mod'));
  if (fs.existsSync(path.join(scratch, 'go.sum'))) fs.copyFileSync(path.join(scratch, 'go.sum'), path.join(evidence, 'consumer-go.sum'));
  for (const round of [1, 2]) {
    env.TEAQL_PARTIAL_ROOT_ROUND = String(round);
    const log = await run([binary, '-test.v', '-test.count=1', '-test.run', '^TestGenerated'], 'round-' + round + '.log', 90000);
    for (const marker of ['FIX_E_EXAMPLE=', 'CORE_LOAD_EXAMPLE=', 'CORE_MUTATION_OWNERSHIP=', 'CORE_MUTATION_PAGING=', 'CORE_VERSION_LIFECYCLE=', 'GO_CHECKER_METADATA=', 'GO_CHECKER_EXACT_RANGES=', 'GO_CONTEXT_ROOT_RESULT=', 'GO_PARTIAL_ANCESTORS_RESULT=', 'GO_PARTIAL_ROOT_RESULT=', 'GO_CHECKER_BOUNDARY_RESULT=']) {
      const values = log.split(/\r?\n/).filter(line => line.startsWith(marker)).map(line => JSON.parse(line.slice(marker.length)));
      assert.equal(values.length, 1, marker); assert.equal(values[0].passed, true, marker);
    }
    const boundaries = log.split(/\r?\n/).filter(line => line.startsWith('GO_CHECKER_BOUNDARY_CASE='))
      .map(line => JSON.parse(line.slice('GO_CHECKER_BOUNDARY_CASE='.length)));
    assert.equal(boundaries.length, 28); assert(boundaries.every(v => v.passed && v.round === round));
    const projections = log.split(/\r?\n/).filter(line => line.startsWith('GO_PROJECTION_VIEW_CASE='))
      .map(line => JSON.parse(line.slice('GO_PROJECTION_VIEW_CASE='.length)));
    assert.equal(projections.length, 3);
    assert.deepEqual(projections.map(v => v.case).sort(), ['independent-query-projection-boundary',
      'same-identity-independent-mutation-ledgers', 'intra-query-repeated-identity-projection-boundary'].sort());
    assert(projections.every(v => v.passed && v.round === round));
    const unchanged = log.split(/\r?\n/).filter(line => line.startsWith('GO_UNCHANGED_VIEW_CASE='))
      .map(line => JSON.parse(line.slice('GO_UNCHANGED_VIEW_CASE='.length)));
    assert.deepEqual(unchanged.map(v => v.case).sort(), ['sparse-restored-values-null-properties-and-edge',
      'repeat-clean-sparse-save', 'other-row-pending-intent-and-values-retained',
      'row-owned-properties-and-loaded-null-retained'].sort());
    assert(unchanged.every(v => v.passed && v.round === round && Object.values(v.effects).every(n => n === 0)));
    const writes = log.split(/\r?\n/).filter(line => line.startsWith('GO_UNCHANGED_VIEW_WRITE='))
      .map(line => JSON.parse(line.slice('GO_UNCHANGED_VIEW_WRITE='.length)));
    assert.equal(writes.length, 1);
    assert(writes[0].passed && writes[0].round === round && writes[0].version === 2 && writes[0].staleProperties === 0);
    const rollback = log.split(/\r?\n/).filter(line => line.startsWith('GO_ROLLBACK_VIEW_CASE='))
      .map(line => JSON.parse(line.slice('GO_ROLLBACK_VIEW_CASE='.length)));
    const retries = log.split(/\r?\n/).filter(line => line.startsWith('GO_ROLLBACK_VIEW_RETRY='))
      .map(line => JSON.parse(line.slice('GO_ROLLBACK_VIEW_RETRY='.length)));
    assert.deepEqual(rollback.map(v => v.case), ['after-write', 'before-commit']);
    assert.deepEqual(retries.map(v => v.case), ['after-write', 'before-commit']);
    assert(rollback.every(v => v.passed && v.round === round && v.queryPropertiesPreserved && v.selectedNullPresent));
    assert(retries.every(v => v.passed && v.round === round && v.version === 2 && v.staleProperties === 0));
    report.expressionBoundaryRounds.push(validateExpressionBoundaries(log));
    assert.deepEqual(fingerprint(), m);
  }
  report.databaseSha256 = sha(fs.readFileSync(path.join(evidence, 'fixture.sqlite'))); report.passed = true;
  report.contextRootDatabase = {file: 'fixture.sqlite.context-root.sqlite',
    sha256: sha(fs.readFileSync(path.join(evidence, 'fixture.sqlite.context-root.sqlite')))};
  report.checkerBoundaryDatabase = {file: 'fixture.sqlite.checker-boundaries.sqlite',
    sha256: sha(fs.readFileSync(path.join(evidence, 'fixture.sqlite.checker-boundaries.sqlite')))};
  report.expressionBoundaryDatabase = {file: 'fixture.sqlite.expression-boundaries.sqlite',
    sha256: sha(fs.readFileSync(path.join(evidence, 'fixture.sqlite.expression-boundaries.sqlite')))};
  report.partialAncestorDatabase = {file: 'fixture.sqlite.partial-ancestors.sqlite',
    sha256: sha(fs.readFileSync(path.join(evidence, 'fixture.sqlite.partial-ancestors.sqlite')))};
  report.partialRootDatabase = {file: 'fixture.sqlite.partial-root.sqlite',
    sha256: sha(fs.readFileSync(path.join(evidence, 'fixture.sqlite.partial-root.sqlite')))};
  report.projectionDatabase = {file: 'fixture.sqlite.projection-views.sqlite',
    sha256: sha(fs.readFileSync(path.join(evidence, 'fixture.sqlite.projection-views.sqlite')))};
  report.unchangedDatabase = {file: 'fixture.sqlite.unchanged.projection-views.sqlite',
    sha256: sha(fs.readFileSync(path.join(evidence, 'fixture.sqlite.unchanged.projection-views.sqlite')))};
  console.log('PASS Go retained Fix/E/loading/ownership example twice on one database');
} catch (error) { console.error(error.stack); process.exitCode ||= 1; }
finally {
  report.disposedOwnedBytes = runner.safeToClean ? ownedBytes(scratch) : 0;
  if (runner.safeToClean) fs.rmSync(scratch, {recursive: true});
  else report.unsafeBuildRetention = {path: scratch, processGroup: runner.unsafeGroup};
  report.scratchRemoved = !fs.existsSync(scratch); runner.dispose();
  fs.writeFileSync(path.join(evidence, 'receipt.json'), JSON.stringify(report, null, 2) + '\n');
  console.log('Retained compact example evidence: ' + evidence);
}
}
main().catch(error => { console.error(error.stack); process.exitCode ||= 1; });
