import assert from 'node:assert/strict';
import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { addCore, loadStore, parseAddress, parsePort, removeCore, saveStore } from '../lib/cores.mjs';

const dir = await mkdtemp(join(tmpdir(), 'nvx-mng-test-'));
const path = join(dir, 'cores.json');
const lab = { alias: 'lab', address: '10.0.0.5:8090', protocol: 'grpc', secretKey: 'k' };

try {
  assert.deepEqual(await loadStore(path), { cores: {} });

  await saveStore(path, addCore(await loadStore(path), lab));
  assert.equal((await loadStore(path)).cores.lab.address, '10.0.0.5:8090');

  const stored = await loadStore(path);
  assert.throws(() => addCore(stored, lab), /already exists/);
  assert.deepEqual(removeCore(await loadStore(path), 'lab').cores, {});
  assert.throws(() => removeCore({ cores: {} }, 'missing'), /no such core/);

  assert.throws(() => addCore({ cores: {} }, { ...lab, protocol: 'soap' }), /protocol/);
  assert.throws(() => addCore({ cores: {} }, { ...lab, address: '10.0.0.5' }), /host:port/);

  assert.deepEqual(parseAddress('10.0.0.5:8090'), { host: '10.0.0.5', port: 8090 });
  assert.equal(parsePort('8181'), 8181);
  assert.throws(() => parsePort('0'), /invalid port/);
} finally {
  await rm(dir, { recursive: true, force: true });
}

console.log('smoke ok');
