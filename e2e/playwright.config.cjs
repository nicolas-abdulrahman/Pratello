const { defineConfig } = require('@playwright/test');
const { execSync } = require('child_process');

function getChromiumExecutable() {
  if (process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE) {
    return process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE;
  }
  try {
    const found = execSync('which google-chrome-stable || which google-chrome || which chromium || true', {
      encoding: 'utf-8',
      stdio: ['pipe', 'pipe', 'ignore'],
    }).trim();
    if (found) return found;
  } catch {}
  return undefined;
}

const baseURL = 'http://127.0.0.1:5177';
module.exports = defineConfig({
  testDir: '.',
  use: {
    baseURL,
    trace: 'retain-on-failure',
    launchOptions: { executablePath: getChromiumExecutable() },
  },
  webServer: {
    command: 'npm --prefix ../frontend run dev -- --host 127.0.0.1 --port 5177 --strictPort',
    url: baseURL,
  },
});
