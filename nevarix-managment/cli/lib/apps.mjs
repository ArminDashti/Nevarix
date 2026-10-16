import path from 'node:path';
import { fileURLToPath } from 'node:url';

// The app source is this repo: cli/lib -> cli -> repo root.
const sourceDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');

// Descriptor for the managment dashboard. Docker compose file is the source of truth for ports and env.
export const APP = {
  name: 'managment',
  folder: 'nevarix-managment',
  image: 'nevarix-webui:latest',
  compose: 'docker-compose.yml',
  dockerfile: 'dockerfile',
  web: { output: 'dist', basePath: '/nevarix/' },
  defaultPort: 5184,
  portEnv: 'PUBLISH_PORT',
  urlPath: '/nevarix/',
  nativeEnv: () => ({}),
  sourceDir,
};

export const APPS = { managment: APP };
