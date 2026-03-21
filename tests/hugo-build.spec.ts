import { test, expect } from '@playwright/test';
import { execFileSync } from 'child_process';
import { existsSync } from 'fs';

test.describe('Hugo Build', () => {
  test('hugo --minify exits successfully @smoke', () => {
    const result = execFileSync('hugo', ['--minify'], { encoding: 'utf-8', timeout: 30_000 });
    expect(result).toBeDefined();
  });

  test('public/index.html exists after build @smoke', () => {
    execFileSync('hugo', ['--minify'], { timeout: 30_000 });
    expect(existsSync('public/index.html')).toBe(true);
  });
});
