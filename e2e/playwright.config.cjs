const { defineConfig } = require('@playwright/test');

const baseURL = 'http://127.0.0.1:5177';
module.exports = defineConfig({
  testDir: '.',
  use: {
    baseURL,
    trace: 'retain-on-failure',
    launchOptions: { executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE },
  },
  webServer: {
    command: 'npm --prefix ../frontend run dev -- --host 127.0.0.1 --port 5177 --strictPort',
    url: baseURL,
  },
});
