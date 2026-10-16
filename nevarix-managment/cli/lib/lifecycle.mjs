import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { createInterface } from 'node:readline/promises';
import { APPS } from './apps.mjs';
import { appPaths, loadConfig, parsePort, saveConfig } from './config.mjs';
import * as docker from './docker.mjs';
import * as native from './native.mjs';
import { capture, portFree, probe, runStep, waitUntil } from './proc.mjs';

export const EXIT = { ok: 0, failed: 1, stopped: 3, notInstalled: 4 };

async function load(name) {
  const app = APPS[name];
  const cfg = await loadConfig();
  const s = cfg.apps[name] ?? {};
  return {
    app,
    cfg,
    s,
    mode: s.mode ?? 'docker',
    port: s.port ?? app.defaultPort,
    paths: appPaths(name),
  };
}

async function requireInstalled(name) {
  const ctx = await load(name);
  if (!ctx.s.installed) throw new Error(`${name} is not installed; run: nevarix-managment rebuild`);
  return ctx;
}

async function save(ctx, patch) {
  ctx.cfg.apps[ctx.app.name] = { ...ctx.s, ...patch };
  ctx.s = ctx.cfg.apps[ctx.app.name];
  await saveConfig(ctx.cfg);
}

function urlOf(ctx, port = ctx.port) {
  return port && ctx.app.urlPath ? `http://127.0.0.1:${port}${ctx.app.urlPath}` : null;
}

function isRunning(ctx) {
  return ctx.mode === 'docker' ? docker.running(ctx.app, ctx.port) : native.isRunning(ctx.app.name);
}

async function startApp(ctx, port) {
  if (port && !(await portFree(port))) throw new Error(`port ${port} is already in use`);
  if (ctx.mode === 'docker') docker.up(ctx.app, port);
  else await native.start(ctx.app, port);
  const url = urlOf(ctx, port);
  if (url && !(await waitUntil(() => probe(url), 30))) throw new Error(`${url} not reachable after 30s`);
}

async function stopApp(ctx) {
  if (ctx.mode === 'docker') docker.stop(ctx.app, ctx.port);
  else await native.stop(ctx.app.name);
  if (await isRunning(ctx)) throw new Error(`${ctx.app.name} still running after stop`);
  if (ctx.port && !(await waitUntil(() => portFree(ctx.port), 30))) {
    throw new Error(`port ${ctx.port} still busy after stop`);
  }
}

async function sourceVersion(app) {
  try {
    return JSON.parse(await readFile(path.join(app.sourceDir, 'package.json'), 'utf8')).version;
  } catch {}
  const g = capture('git', ['describe', '--tags', '--always', '--dirty'], { cwd: app.sourceDir });
  return g.ok ? g.stdout : 'unknown';
}

async function confirm(question, yes) {
  if (yes) return true;
  const rl = createInterface({ input: process.stdin, output: process.stdout });
  try {
    return /^y(es)?$/i.test((await rl.question(`${question} [y/N] `)).trim());
  } finally {
    rl.close();
  }
}

function wipeTargets(ctx) {
  const what = ctx.mode === 'docker'
    ? `containers, networks, volumes of project nevarix-${ctx.app.name}, image ${ctx.app.image}`
    : `${ctx.paths.base} (app, data, logs)`;
  return `${what}, and config entry`;
}

export async function status(name) {
  const ctx = await load(name);
  if (!ctx.s.installed) {
    console.log(`${name}: not installed`);
    return EXIT.notInstalled;
  }
  const running = await isRunning(ctx);
  const url = urlOf(ctx);
  const reachable = url && running ? await probe(url) : null;
  console.log(`${name}`);
  console.log(`  mode:     ${ctx.mode}`);
  console.log(`  version:  ${ctx.s.version ?? 'unknown'}`);
  console.log(`  port:     ${ctx.port ?? '-'}`);
  console.log(`  running:  ${running ? 'yes' : 'no'}`);
  console.log(`  url:      ${url ?? '-'}${reachable === null ? '' : reachable ? ' (reachable)' : ' (NOT reachable)'}`);
  console.log(`  paths:    ${ctx.mode === 'docker' ? ctx.app.sourceDir : ctx.paths.install}`);
  return running ? EXIT.ok : EXIT.stopped;
}

export async function version(name) {
  const ctx = await load(name);
  const source = await sourceVersion(ctx.app);
  console.log(`${name} installed: ${ctx.s.version ?? 'not installed'}`);
  console.log(`${name} source:    ${source}`);
  if (ctx.s.version && ctx.s.version !== source) {
    console.log('MISMATCH: source changed since install; run "rebuild" or "update"');
    return EXIT.failed;
  }
  return EXIT.ok;
}

export async function run(name, { port } = {}) {
  const ctx = await requireInstalled(name);
  const p = port ? parsePort(port) : ctx.port;
  if (await isRunning(ctx)) {
    console.log(`${name} already running`);
    return status(name);
  }
  await startApp({ ...ctx, port: p }, p);
  if (p !== ctx.port) await save(ctx, { port: p });
  return status(name);
}

export async function stop(name) {
  const ctx = await load(name);
  if (!ctx.s.installed) {
    console.log(`${name} is not installed; nothing to stop`);
    return EXIT.ok;
  }
  await stopApp(ctx);
  console.log(`${name} stopped`);
  return EXIT.ok;
}

export async function setPort(name, value) {
  const ctx = await requireInstalled(name);
  if (value === undefined) {
    console.log(`${name} port: ${ctx.port ?? 'none'}`);
    return EXIT.ok;
  }
  const p = parsePort(value);
  const running = await isRunning(ctx);
  if (running) await stopApp(ctx);
  await save(ctx, { port: p });
  if (running) await startApp({ ...ctx, port: p }, p);
  return status(name);
}

export async function rebuild(name, { removeData = 'no', yes = false, mode } = {}) {
  const ctx = await load(name);
  const nextMode = mode ?? ctx.mode;
  const wasRunning = ctx.s.installed && (await isRunning(ctx));
  if (removeData === 'yes') {
    const ok = await confirm(`Wipe ALL data and config for ${name}: ${wipeTargets(ctx)}?`, yes);
    if (!ok) {
      console.log('aborted');
      return EXIT.failed;
    }
  }
  console.log(`rebuild ${name} (mode ${nextMode}, remove-data ${removeData})`);

  if (ctx.s.installed) await stopApp(ctx);
  if (nextMode === 'docker') {
    if (ctx.s.installed && ctx.mode === 'docker') docker.down(ctx.app, { volumes: removeData === 'yes' });
    docker.build(ctx.app);
  } else {
    await native.build(ctx.app, ctx.paths.staging);
    if (removeData === 'yes') await native.wipe(name);
    await native.swap(name);
  }

  const base = removeData === 'yes' ? {} : ctx.s;
  ctx.cfg.apps[name] = {
    ...base,
    installed: true,
    mode: nextMode,
    port: base.port ?? ctx.app.defaultPort,
    version: await sourceVersion(ctx.app),
  };
  ctx.s = ctx.cfg.apps[name];
  await saveConfig(ctx.cfg);

  if (wasRunning) await startApp({ ...ctx, mode: nextMode, port: ctx.s.port }, ctx.s.port);
  return status(name);
}

export async function update(name) {
  const ctx = await requireInstalled(name);
  runStep(`git pull --ff-only (${name})`, 'git', ['pull', '--ff-only'], { cwd: ctx.app.sourceDir });
  return rebuild(name, { removeData: 'no' });
}

export async function uninstall(name, { yes = false } = {}) {
  const ctx = await load(name);
  if (!ctx.s.installed) {
    console.log(`${name} is not installed`);
    return EXIT.ok;
  }
  const ok = await confirm(`Remove ${name}: ${wipeTargets(ctx)}?`, yes);
  if (!ok) return EXIT.failed;
  if (ctx.mode === 'docker') docker.down(ctx.app, { volumes: true });
  else await native.stop(name);
  if (ctx.mode === 'native') await native.wipe(name);
  delete ctx.cfg.apps[name];
  await saveConfig(ctx.cfg);
  console.log(`${name} removed`);
  return EXIT.ok;
}