import { spawnSync } from 'node:child_process';
import { createConnection } from 'node:net';
import { setTimeout as delay } from 'node:timers/promises';

const port = 6379;
const container = 'pratello-dev-redis';

function redisReady() {
  return new Promise((resolve) => {
    const socket = createConnection({ host: 'localhost', port });
    socket.setTimeout(500);
    let reply = '';

    function finish(ready) {
      socket.destroy();
      resolve(ready);
    }

    socket.on('connect', () => socket.write('*1\r\n$4\r\nPING\r\n'));
    socket.on('data', (chunk) => {
      reply += chunk.toString();
      if (reply.includes('\r\n')) finish(reply === '+PONG\r\n');
    });
    socket.once('error', () => finish(false));
    socket.once('timeout', () => finish(false));
    socket.once('close', () => finish(false));
  });
}

if (await redisReady()) {
  console.log('Using Redis at redis://localhost:6379/0');
} else {
  const started = spawnSync('docker', [
    'run', '--detach', '--rm', '--name', container,
    '--publish', `127.0.0.1:${port}:6379`, 'redis:7-alpine',
  ], { stdio: 'inherit' });

  if (started.status !== 0) {
    console.error('Could not start development Redis. Start Docker and make sure port 6379 is available.');
    process.exitCode = 1;
  } else {
    let ready = false;
    for (let attempt = 0; attempt < 40 && !ready; attempt++) {
      ready = await redisReady();
      if (!ready) await delay(250);
    }
    if (ready) {
      console.log('Development Redis ready at redis://localhost:6379/0');
    } else {
      console.error(`Development Redis did not become ready. Check: docker logs ${container}`);
      process.exitCode = 1;
    }
  }
}
