import { spawn } from 'node:child_process';
import { readFile } from 'node:fs/promises';
import { createInterface } from 'node:readline';

const modules = ['IDP', 'TEN', 'CAT', 'ORD', 'FIN', 'REV'];
const prefix = modules.join('|');
const docs = await readFile('docs/Casos de teste.md', 'utf8');
const cases = new Map([...docs.matchAll(new RegExp(`^(?:### |\\| )(${prefix})-(\\d{3})(.*)$`, 'gm'))]
  .map(([, module, number, description]) => [module + number, { planned: /planejado/i.test(description), tests: new Set() }]));
const grouped = new Map();
const newEvidence = {
  TestMenuPostgresLifecycle: 'CAT001 CAT003 CAT006 CAT009 CAT010 CAT013 CAT020 CAT023 CAT024 CAT026 CAT027 CAT028 CAT029 CAT030 CAT034 CAT036 CAT037 CAT038 CAT039 CAT040 CAT041 CAT042',
  TestOrderQRCodeCheckoutAndAnalyticsPostgres: 'ORD003 ORD024 ORD036',
  TestPostgresPaymentLifecycle: 'FIN001 FIN002 FIN005',
};
for (const [test, ids] of Object.entries(newEvidence)) {
  for (const id of ids.split(' ')) grouped.set(id, [test]);
}
const standalone = new Set(['CAT050', 'CAT051', 'FIN025', 'FIN026', 'FIN027']);
for (const [, module, ranges, evidence] of docs.matchAll(/^\| (CAT|ORD|FIN) \| (`[^\n]+?) \| ([^\n]+) \|$/gm)) {
  const tests = [...evidence.matchAll(/`(Test\w+)`/g)].map(([, name]) => name);
  for (const [, first, last] of ranges.matchAll(/\w+-(\d{3})(?:–(\d{3}))?/g)) {
    for (let n = Number(first); n <= Number(last ?? first); n++) grouped.set(module + String(n).padStart(3, '0'), tests);
  }
}

if (!process.env.AUTH_TEST_DATABASE_URL && process.env.DATABASE_URL) {
  process.env.AUTH_TEST_DATABASE_URL = process.env.DATABASE_URL;
}
if (!process.env.AUTH_TEST_REDIS_URL && process.env.REDIS_URL) {
  process.env.AUTH_TEST_REDIS_URL = process.env.REDIS_URL;
}

const child = spawn('go', ['test', '-json', '-count=1', './server/...'], { stdio: ['ignore', 'pipe', 'inherit'] });
const exit = new Promise(resolve => {
  child.on('close', resolve);
  child.on('error', error => { console.error(error.message); resolve(1); });
});
const results = new Map();
const logs = new Map();
const packageStatus = new Map();
const rank = { PASS: 0, XFAIL: 1, SKIP: 2, XPASS: 3, FAIL: 4 };
const expectations = new Map();
console.log('Running backend Go tests…');
for await (const line of createInterface({ input: child.stdout })) {
  const { Action, Package, Test, Output } = JSON.parse(line);
  if (Output && Package) {
    if (!logs.has(Package)) logs.set(Package, []);
    logs.get(Package).push(Output);
  }
  if (Action === 'output' && Test && !Test.includes('/')) {
    const marker = Output?.match(/^\s+\S+\.go:\d+: (XFAIL|XPASS): /);
    if (marker) expectations.set(Test, marker[1]);
  }
  if (!['pass', 'skip', 'fail'].includes(Action)) continue;
  const status = Action.toUpperCase();
  if (!Test) { packageStatus.set(Package, status); continue; }
  const expected = expectations.get(Test);
  results.set(Test, (status === 'PASS' && expected === 'XFAIL') ||
    (status === 'FAIL' && expected === 'XPASS') ? expected : status);
  const [name, subtest] = Test.split('/');
  const numbered = name.match(new RegExp(`^Test(${prefix})(\\d{3}(?:(?:And)?\\d{3})*)`));
  if (numbered && !subtest) {
    for (const number of numbered[2].match(/\d{3}/g)) cases.get(numbered[1] + number)?.tests.add(name);
  }
}
const code = await exit;
for (const [pkg, status] of packageStatus) {
  if (status === 'FAIL') process.stderr.write((logs.get(pkg) ?? []).join(''));
}

function outcome(test) {
  const statuses = [...results].filter(([name]) => name === test || name.startsWith(`${test}/`)).map(([, status]) => status);
  return statuses.reduce((worst, status) => rank[status] > rank[worst] ? status : worst, results.get(test) ?? 'SKIP');
}
const counts = { PASS: 0, XFAIL: 0, XPASS: 0, FAIL: 0, SKIP: 0, UNVERIFIED: 0, PLANNED: 0 };
const colors = { PASS: 32, XFAIL: 33, XPASS: 31, FAIL: 31, SKIP: 33, UNVERIFIED: 33, PLANNED: 90 };
for (const module of modules) {
  console.log(`\n── ${module} ${'─'.repeat(24)}`);
  for (const [id, entry] of [...cases].sort(([a], [b]) => a.localeCompare(b))) {
    if (!id.startsWith(module)) continue;
    const suites = standalone.has(id) ? undefined : grouped.get(id);
    const tests = new Set([...entry.tests, ...(suites ?? [])]);
    const status = entry.planned ? 'PLANNED' : tests.size === 0 ? 'UNVERIFIED' :
      [...tests].map(outcome).reduce((worst, status) => rank[status] > rank[worst] ? status : worst, 'PASS');
    counts[status]++;
    const display = process.stdout.isTTY && !process.env.NO_COLOR ? `\x1b[${colors[status]}m${status}\x1b[0m` : status;
    console.log(`  ${id}  ──────────  ${display}${suites ? '  (shared suite)' : ''}`);
  }
}
console.log(`\nBackend cases executed: ${counts.PASS + counts.XFAIL + counts.XPASS + counts.FAIL}  │  ${counts.PASS} PASS  │  ${counts.XFAIL} XFAIL  │  ${counts.XPASS} XPASS  │  ${counts.FAIL} FAIL`);
console.log(`Not executed: ${counts.SKIP} skipped  │  ${counts.PLANNED} planned${counts.UNVERIFIED ? `  │  ${counts.UNVERIFIED} missing evidence` : ''}`);
process.exitCode = code === 0 && counts.FAIL === 0 && counts.XPASS === 0 && counts.UNVERIFIED === 0 ? 0 : 1;
