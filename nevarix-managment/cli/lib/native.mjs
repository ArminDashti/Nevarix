import { spawn } from 'node:child_process';
import { existsSync } from 'node:fs';
import { cp, mkdir, open, readFile, rename, rm, stat, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { appPaths } from './config.mjs';
import { isAlive, killTree, runStep } from './proc.mjs';

const isWin = process.platform === 'win32';
const here = path.dirname(fileURLToPath(import.meta.url));
const LOG_LIMIT = 10 * 1024 * 1024;
const exeName = (binary) => binary + (isWin ? '.exe' : '');

export async function build(app, staging) {
  await rm(staging, { recursive: true, force: true });
  if (app.go) {
    const bin = path.join(staging, 'bin');
    await mkdir(bin, { recursive: true });
    runStep(`go build ${app.name}`, 'go', ['build', '-o', path.join(bin, exeName(app.go.binary)), app.go.pkg], {
      cwd: app.sourceDir,
      env: { CGO_ENABLED: '0' },
    });
    return;
  }
  if (!existsSync(path.join(app.sourceDir, 'node_modules'))) {
    throw new Error(`${app.folder}: node_modules missing; run "npm ci" there first (the CLI does not install packages)`);
  }
  runStep(`npm run build ${app.name}`, 'npm', ['run', 'build'], {
    cwd: app.sourceDir,
    env: { VITE_BASE_PATH: app.web.basePath },
  });
  await cp(path.join(app.sourceDir, app.web.output), path.join(staging, 'web'), { recursive: true });
}

// Old install is moved away only after staging is complete, so a failed build keeps the old app.
export async function swap(name) {
  const p = appPaths(name);
  await mkdir(p.base, { recursive: true });
  await rm(p.install, { recursive: true, force: true });
  await rename(path.join(p.staging), p.install);
}

export async function wipe(name) {
  await rm(appPaths(name).base, { recursive: true, force: true });
}

async function readPid(name) {
  try {
    return Number((await readFile(appPaths(name).pid, 'utf8')).trim()) || null;
  } catch {
    return null;
  }
}

export async function isRunning(name) {
  const pid = await readPid(name);
  return pid !== null && isAlive(pid);
}

export async function start(app, port) {
  const p = appPaths(app.name);
  await mkdir(p.logs, { recursive: true });
  await mkdir(p.data, { recursive: true });
  const log = path.join(p.logs, `${app.name}.log`);
  await rotate(log);
  const out = await open(log, 'a');
  const [cmd, args] = app.go
    ? [path.join(p.install, 'bin', exeName(app.go.binary)), []]
    : [process.execPath, [path.join(here, 'static-server.mjs'), path.join(p.install, 'web'), String(port), app.web.basePath]];
  const child = spawn(cmd, args, {
    cwd: p.install,
    detached: true,
    stdio: ['ignore', out.fd, out.fd],
    windowsHide: true,
    env: { ...process.env, ...app.nativeEnv(port), APP_DATA_DIR: p.data },
  });
  try {
    await new Promise((resolve, reject) => {
      child.once('error', reject);
      child.once('spawn', resolve);
    });
  } finally {
    await out.close();
  }
  child.unref();
  await writeFile(p.pid, String(child.pid));
}

export async function stop(name) {
  const pid = await readPid(name);
  if (pid !== null && isAlive(pid)) await killTree(pid);
  await rm(appPaths(name).pid, { force: true });
}

async function rotate(log) {
  try {
    if ((await stat(log)).size > LOG_LIMIT) await rename(log, `${log}.1`);
  } catch (err) {
    if (err.code !== 'ENOENT') throw err;
  }
}
