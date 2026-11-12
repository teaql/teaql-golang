'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path');
const requested=process.env.TEAQL_NATURAL_MESSAGES_EVIDENCE_DIR;
// The main verifier owns process groups and disposable builds. Compact evidence
// and its isolated database are retained; no generated source is modified.
process.argv.push('--example-only','--output',requested?path.resolve(requested):path.join(fs.mkdtempSync(path.join(os.tmpdir(),'teaql-go-natural-evidence-')),'run'));
require('../../scripts/verify-natural-messages.cjs');
