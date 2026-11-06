import { defineConfig } from '@playwright/test';

const puerto = process.env.E2E_BASE_URL ?? 'http://localhost:8099';

export default defineConfig({
  testDir: '.',
  fullyParallel: false,
  workers: 1,
  timeout: 30_000,
  expect: { timeout: 5_000 },
  reporter: [['list']],
  use: {
    baseURL: puerto,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    locale: 'es-PE',
  },
});
