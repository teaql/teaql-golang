'use strict';
const assert = require('node:assert/strict');
const NAMES = ['late-leaf', 'indexed-leaf', 'loaded-null', 'loaded-empty',
  'out-of-range', 'terminal-not-loaded', 'sibling-paths', 'fk-only-not-loaded',
  'no-side-effects', 'native-location-cases'];
const ZERO = {statements: 0, transactions: 0, policies: 0, audits: 0};
const parse = (log, prefix) => log.split(/\r?\n/).filter(s => s.startsWith(prefix)).map(s => JSON.parse(s.slice(prefix.length)));
const segments = values => values.map(v => typeof v === 'number' ? {kind: 'index', index: v} : {kind: 'property', name: v});
function failure(value, location, attempted, pointer, index, chain, native, attemptedPointer = pointer) {
  const wire = value.wire;
  assert.deepEqual(Object.keys(wire).sort(), ['code', 'location', 'instancePath', 'breakPoint',
    'attemptedLocation', 'attemptedInstancePath', 'breakPointIndex', 'originalExpression',
    'expressionWithBrokenPoint', 'suggestedFix'].sort());
  assert.equal(wire.code, 'NOT_LOADED');
  assert.deepEqual(wire.location, segments(location));
  assert.deepEqual(wire.attemptedLocation, segments(attempted));
  assert.equal(wire.instancePath, pointer);
  assert.equal(wire.attemptedInstancePath, attemptedPointer);
  assert.equal(wire.breakPointIndex, index);
  assert.equal(wire.breakPoint, location.at(-1));
  assert.equal(wire.expressionWithBrokenPoint, chain);
  assert.equal(wire.originalExpression, chain.replace('<broken>', ''));
  assert.equal(chain.split('<broken>').length, 2);
  assert.deepEqual(wire.suggestedFix, {action: 'preload', location: segments(location)});
  assert.equal(value.human, 'TeaQLNotLoadedError: access_path=' + native + ' break_point=' + location.at(-1)
    + ' — select the missing field or relation expression=' + chain);
  assert(!/PRIVATE_GO_|\bid=|"instancePath"/.test(value.human));
}
function validateExpressionBoundaries(log) {
  assert(!/--- FAIL:|--- SKIP:|DATA RACE|panic:/m.test(log));
  assert.equal((log.match(/^--- PASS: TestGeneratedExpressionBoundaries /gm) || []).length, 1);
  const results = parse(log, 'CORE_GO_E_BOUNDARY_RESULT=');
  assert.equal(results.length, 1);
  const result = results[0], cases = parse(log, 'CORE_GO_E_CASE=');
  assert.equal(result.language, 'go'); assert.equal(result.passed, true);
  assert.equal(result.generatedCases, 9); assert.equal(result.nativeLocationCases, 2);
  assert.equal(result.terminalObservations, 3);
  assert.equal(result.physicalVersionsUnchanged, true);
  assert.equal(result.fullPunctuationGeneratedPathAccepted, false);
  assert.deepEqual(result.effects, ZERO);
  assert.deepEqual(cases.map(v => v.name), NAMES);
  assert.deepEqual(result.cases, NAMES);
  const rows = Object.fromEntries(cases.map(v => [v.name, v.value]));
  failure(rows['late-leaf'], ['tenant', 'name'], ['tenant', 'name'], '/tenant/name', 1,
    'E.PurchaseOrder(<root>).Tenant().Name()<broken>.TryEval()', 'Tenant.Name');
  failure(rows['indexed-leaf'], ['order_item_list', 2, 'product_name'], ['order_item_list', 2, 'product_name'],
    '/orderItemList/2/productName', 2, 'E.PurchaseOrder(<root>).OrderItemList().Get(2).ProductName()<broken>.TryEval()', 'OrderItemList[2].ProductName');
  assert.deepEqual(rows['loaded-null'], {value: null, present: false, fallbackUsed: true});
  assert.deepEqual(rows['loaded-empty'], {size: 0, value: null, present: false, fallbackUsed: true});
  assert.deepEqual(rows['out-of-range'], {index: 3, negativeIndex: -1, value: null, present: false, fallbackUsed: true});
  const terminals = rows['terminal-not-loaded'].terminals;
  assert.equal(terminals.length, 3);
  for (const [i, terminal] of ['TryEval()', 'Eval()', 'OrElse(<fallback>)'].entries()) {
    failure(terminals[i], ['tenant'], ['tenant', 'name'], '/tenant', 0,
      'E.PurchaseOrder(<root>).Tenant()<broken>.Name().' + terminal, 'Tenant', '/tenant/name');
  }
  const sibling = rows['sibling-paths'];
  assert.deepEqual(sibling.leftBefore, sibling.leftAfter);
  failure(sibling.leftBefore, ['tenant'], ['tenant', 'name'], '/tenant', 0,
    'E.PurchaseOrder(<root>).Tenant()<broken>.Name().TryEval()', 'Tenant', '/tenant/name');
  failure(sibling.right, ['order_item_list'], ['order_item_list', 2, 'product_name'], '/orderItemList', 0,
    'E.PurchaseOrder(<root>).OrderItemList()<broken>.Get(2).ProductName().TryEval()', 'OrderItemList', '/orderItemList/2/productName');
  assert.equal(rows['fk-only-not-loaded'].fk, 1);
  failure(rows['fk-only-not-loaded'].failure, ['tenant'], ['tenant', 'name'], '/tenant', 0,
    'E.PurchaseOrder(<root>).Tenant()<broken>.Name().TryEval()', 'Tenant', '/tenant/name');
  const pure = rows['no-side-effects'];
  assert.match(pure.before, /^[a-f0-9]{64}$/); assert.equal(pure.after, pure.before);
  assert.deepEqual(pure.effects, ZERO); assert.equal(pure.ledgerIdentitiesUnchanged, true);
  assert.deepEqual(Object.keys(pure.calibration).sort(), Object.keys(ZERO).sort());
  assert(Object.values(pure.calibration).every(n => Number.isSafeInteger(n) && n > 0));
  const native = rows['native-location-cases'];
  assert.equal(native.nativeOnly, true);
  assert.deepEqual(native.profiles.map(v => v.profile), ['camelCase', 'snake_case', 'PascalCase']);
  for (const p of native.profiles) {
    assert.equal(p.punctuation.instancePath, p.profile === 'PascalCase' ? '/A~0~1b.c' : '/a~0~1b.c');
    assert.deepEqual(p.punctuation.location, segments(['a~/b.c']));
    assert.equal(p.repeated.expressionWithBrokenPoint, 'E.Native(<root>).Tenant()<broken>.Tenant().Name().TryEval()');
    assert.equal(p.repeated.expressionWithBrokenPoint.split('<broken>').length, 2);
  }
  assert(!JSON.stringify(cases).includes('PRIVATE_GO_'));
  for (const id of [result.parentId, result.emptyParentId, ...result.childIds]) assert(Number.isSafeInteger(id) && id > 0);
  assert.notEqual(result.parentId, result.emptyParentId);
  assert.equal(result.childIds.length, 3); assert.equal(new Set(result.childIds).size, 3);
  return result;
}
module.exports = {validateExpressionBoundaries};
