import { capture, runStep } from './proc.mjs';

const project = (app) => `nevarix-${app.name}`;
const composeArgs = (app) => ['compose', '-p', project(app), '-f', app.compose];

function env(app, port) {
  return {
    IMAGE_TAG: app.image,
    ...(app.portEnv && port ? { [app.portEnv]: String(port) } : {}),
  };
}

function compose(app, args, port) {
  runStep(`docker compose ${args.join(' ')} (${app.name})`, 'docker', [...composeArgs(app), ...args], {
    cwd: app.sourceDir,
    env: env(app, port),
  });
}

function ensureNetwork() {
  const net = process.env.DOCKER_NETWORK ?? 'nevarix-net';
  if (!capture('docker', ['network', 'inspect', net]).ok) {
    runStep(`create network ${net}`, 'docker', ['network', 'create', net]);
  }
}

export function running(app, port) {
  const r = capture('docker', [...composeArgs(app), 'ps', '-q', '--status', 'running'], {
    cwd: app.sourceDir,
    env: env(app, port),
  });
  return r.ok && r.stdout.length > 0;
}

export function up(app, port) {
  ensureNetwork();
  compose(app, ['up', '-d'], port);
}

export function stop(app, port) {
  compose(app, ['stop'], port);
}

export function down(app, { volumes = false } = {}) {
  compose(app, ['down', '--remove-orphans', ...(volumes ? ['-v', '--rmi', 'local'] : [])]);
}

export function build(app) {
  runStep(`docker build ${app.name}`, 'docker', ['build', '--no-cache', '--pull', '-f', app.dockerfile, '-t', app.image, '.'], {
    cwd: app.sourceDir,
  });
}
