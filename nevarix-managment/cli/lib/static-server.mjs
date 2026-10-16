// Serves a built SPA under basePath with index.html fallback. Used for Native managment.
import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import path from 'node:path';

const [root, port, basePath = '/'] = process.argv.slice(2);
const types = {
  '.html': 'text/html',
  '.js': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.webmanifest': 'application/manifest+json',
};

async function tryRead(file) {
  try {
    return await readFile(file);
  } catch {
    return null;
  }
}

createServer(async (req, res) => {
  const { pathname } = new URL(req.url, 'http://localhost');
  if (!pathname.startsWith(basePath)) {
    res.writeHead(302, { Location: basePath });
    return res.end();
  }
  const rel = decodeURIComponent(pathname.slice(basePath.length)) || 'index.html';
  let file = path.join(root, rel);
  if (!file.startsWith(root)) {
    res.writeHead(403);
    return res.end();
  }
  let body = await tryRead(file);
  if (body === null) {
    file = path.join(root, 'index.html');
    body = await tryRead(file);
  }
  if (body === null) {
    res.writeHead(404);
    return res.end();
  }
  res.writeHead(200, { 'Content-Type': types[path.extname(file)] ?? 'application/octet-stream' });
  res.end(body);
}).listen(Number(port), '127.0.0.1');
