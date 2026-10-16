import { spawnSync } from 'node:child_process';
import http from 'node:http';
import net from 'node:net';

const isWin = process.platform === 'win32';
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

export function runStep(label, cmd, args, { cwd, env } = {}) {
  console.log(`==> ${label}`);
  const r = spawnSync(cmd, args, {
    cwd,
    env: { ...process.env, ...env },
    stdio: 'inherit',
    shell: isWin && cmd === 'npm',
    windowsHide: true,
  });
  if (r.error) throw r.error;
  if (r.status !== 0) throw new Error(`${label} failed (exit ${r.status})`);
}

export function capture(cmd, args, { cwd, env } = {}) {
  const r = spawnSync(cmd, args, {
    cwd,
    env: { ...process.env, ...env },
    encoding: 'utf8',
    windowsHide: true,
  });
  return { ok: !r.error && r.status === 0, stdout: (r.stdout ?? '').trim() };
}

export function isAlive(pid) {
  try {
    process.kill(pid, 0);
    return true;
  } catch (err) {
    return err.code === 'EPERM';
  }
}

// ponytail: TERM then KILL after 10s; pid reuse after exit is not guarded (ceiling: stale pid file).
export async function killTree(pid) {
  if (isWin) {
    spawnSync('taskkill', ['/PID', String(pid), '/T', '/F'], { windowsHide: true });
    return;
  }
  try {
    process.kill(-pid, 'SIGTERM');
  } catch {}
  for (let i = 0; i < 100 && isAlive(pid); i++) await sleep(100);
  if (isAlive(pid)) {
    try {
      process.kill(-pid, 'SIGKILL');
    } catch {}
  }
}

export function portFree(port) {
  return new Promise((resolve) => {
    const server = net.createServer();
    server.once('error', () => resolve(false));
    server.once('listening', () => server.close(() => resolve(true)));
    server.listen(port, '127.0.0.1');
  });
}

export function probe(url, timeoutMs = 2000) {
  return new Promise((resolve) => {
    const req = http.get(url, (res) => {
      res.resume();
      resolve(true);
    });
    req.setTimeout(timeoutMs, () => {
      req.destroy();
      resolve(false);
    });
    req.on('error', () => resolve(false));
  });
}

export async function waitUntil(check, seconds) {
  for (let i = 0; i < seconds; i++) {
    if (await check()) return true;
    await sleep(1000);
  }
  return false;
}
