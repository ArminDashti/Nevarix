import assert from 'node:assert/strict';
import path from 'node:path';
import { APP } from '../lib/apps.mjs';
import { appPaths, parsePort, ROOT } from '../lib/config.mjs';

assert.equal(parsePort('5184'), 5184);
assert.equal(parsePort(1), 1);
assert.equal(parsePort(65535), 65535);
for (const bad of ['0', '65536', '80.5', 'abc', '']) {
  assert.throws(() => parsePort(bad), /invalid port/, `should reject ${bad}`);
}

assert.equal(APP.defaultPort, 5184);
assert.ok(APP.compose && APP.dockerfile, 'compose and dockerfile are set');

// Staging must sit outside the app base, otherwise a --remove-data wipe deletes it.
const p = appPaths('managment');
assert.ok(!p.staging.startsWith(p.base + path.sep), 'staging is outside app base');
assert.ok(p.staging.startsWith(ROOT), 'staging is under nevarix root');

console.log('smoke ok');
