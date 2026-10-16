import { homedir } from 'node:os';
import path from 'node:path';
import { mkdir, readFile, writeFile } from 'node:fs/promises';

const isWin = process.platform === 'win32';

export const ROOT = isWin
  ? path.join(process.env.LOCALAPPDATA ?? path.join(homedir(), 'AppData', 'Local'), 'nevarix')
  : path.join(homedir(), '.local', 'share', 'nevarix');

export const CONFIG_FILE = isWin
  ? path.join(ROOT, 'config.json')
  : path.join(homedir(), '.config', 'nevarix', 'config.json');

export async function loadConfig() {
  try {
    return JSON.parse(await readFile(CONFIG_FILE, 'utf8'));
  } catch (err) {
    if (err.code === 'ENOENT') return { apps: {} };
    throw err;
  }
}

export async function saveConfig(cfg) {
  await mkdir(path.dirname(CONFIG_FILE), { recursive: true });
  await writeFile(CONFIG_FILE, JSON.stringify(cfg, null, 2));
}

// Staging lives outside the app folder so wiping the app never kills an in-progress build.
export function appPaths(name) {
  const base = path.join(ROOT, name);
  return {
    base,
    install: path.join(base, 'app'),
    data: path.join(base, 'data'),
    logs: path.join(base, 'logs'),
    pid: path.join(base, 'run.pid'),
    staging: path.join(ROOT, '.staging', name),
  };
}

export function parsePort(value) {
  const n = Number(value);
  if (!Number.isInteger(n) || n < 1 || n > 65535) throw new Error(`invalid port: ${value}`);
  return n;
}
