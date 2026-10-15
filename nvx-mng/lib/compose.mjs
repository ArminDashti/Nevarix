import { spawn } from 'node:child_process';
import { writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const PROJECT = 'nevarix';

export const composeFile =
  process.env.NVX_COMPOSE_FILE || resolve(here, '..', '..', 'nevarix-core', 'docker-compose.local.yml');

// Runs `docker compose -f <files> -p nevarix <args>` with inherited stdio.
export function compose(args, extraFiles = []) {
  const fileArgs = [composeFile, ...extraFiles].flatMap((file) => ['-f', file]);
  return new Promise((done, fail) => {
    const child = spawn('docker', ['compose', ...fileArgs, '-p', PROJECT, ...args], { stdio: 'inherit' });
    child.on('error', fail);
    child.on('exit', (code) =>
      code === 0 ? done() : fail(new Error(`docker compose ${args.join(' ')} exited with code ${code}`)),
    );
  });
}

// Compose override that republishes the core HTTP port on the host.
export async function writePortOverride(hostPort) {
  const file = join(tmpdir(), `nvx-mng-ports-${hostPort}.yml`);
  await writeFile(
    file,
    [
      'services:',
      '  core:',
      '    ports: !override',
      `      - "${hostPort}:8090"`,
      '      - "50051:50051"',
      '',
    ].join('\n'),
  );
  return file;
}
