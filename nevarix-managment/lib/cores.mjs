import { connect } from 'node:net';
import { mkdir, readFile, rename, writeFile } from 'node:fs/promises';
import { homedir } from 'node:os';
import { dirname, join } from 'node:path';

export const PROTOCOLS = ['restful', 'grpc'];

// ponytail: secret keys are stored in plain text in the JSON file; upgrade to Windows DPAPI.
export function configPath() {
  if (process.env.NVX_MNG_CONFIG) return process.env.NVX_MNG_CONFIG;
  return join(process.env.LOCALAPPDATA || join(homedir(), 'AppData', 'Local'), 'NevarixManagment', 'cores.json');
}

export async function loadStore(path = configPath()) {
  try {
    const store = JSON.parse(await readFile(path, 'utf8'));
    return { cores: store.cores ?? {} };
  } catch (err) {
    if (err.code === 'ENOENT') return { cores: {} };
    throw err;
  }
}

// Write to a temp file, then rename, so a crash cannot truncate the store.
export async function saveStore(path, store) {
  await mkdir(dirname(path), { recursive: true });
  const tmp = `${path}.tmp`;
  await writeFile(tmp, JSON.stringify(store, null, 2), { mode: 0o600 });
  await rename(tmp, path);
}

export function parsePort(value) {
  const port = Number(value);
  if (!Number.isInteger(port) || port < 1 || port > 65535) throw new Error(`invalid port: ${value}`);
  return port;
}

export function parseAddress(address) {
  const idx = address.lastIndexOf(':');
  if (idx <= 0) throw new Error(`address must be host:port, got: ${address}`);
  return { host: address.slice(0, idx), port: parsePort(address.slice(idx + 1)) };
}

// Pure: returns a new store with the core added. Rejects bad input and duplicate aliases.
export function addCore(store, { alias, address, protocol, secretKey }) {
  if (!alias) throw new Error('alias is required');
  if (!PROTOCOLS.includes(protocol)) throw new Error(`protocol must be one of: ${PROTOCOLS.join(', ')}`);
  if (!secretKey) throw new Error('secret-key is required');
  parseAddress(address);
  if (store.cores[alias]) throw new Error(`core already exists: ${alias}`);
  return { cores: { ...store.cores, [alias]: { address, protocol, secretKey } } };
}

export function removeCore(store, alias) {
  if (!store.cores[alias]) throw new Error(`no such core: ${alias}`);
  const { [alias]: _removed, ...cores } = store.cores;
  return { cores };
}

export function probe(host, port, timeoutMs = 1500) {
  return new Promise((done) => {
    const socket = connect({ host, port });
    const finish = (ok) => {
      socket.destroy();
      done(ok);
    };
    socket.setTimeout(timeoutMs, () => finish(false));
    socket.once('connect', () => finish(true));
    socket.once('error', () => finish(false));
  });
}
