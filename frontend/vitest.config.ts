/// <reference types="vitest/config" />
import { defineConfig, mergeConfig } from 'vitest/config'
import viteConfig from './vite.config'

// Vitest 配置文件
// 继承 vite.config.ts 的配置，添加测试专用配置
export default mergeConfig(viteConfig, defineConfig({
  test: {
    // 使用 jsdom 模拟浏览器环境
    environment: 'jsdom',

    // 启用全局测试 API (describe, it, expect 等)
    globals: true,

    // 测试文件匹配模式
    include: ['src/**/*.{test,spec}.{js,ts,jsx,tsx}'],

    // 排除目录
    exclude: ['node_modules', 'dist', 'e2e'],

    // 测试前运行的 setup 文件
    setupFiles: ['./src/test/setup.ts'],

    // CSS 处理 (避免 CSS 导入错误)
    css: true,

    // 覆盖率配置
    coverage: {
      provider: 'v8',
      enabled: false, // 默认不启用，使用 --coverage 手动启用
      reporter: ['text', 'json', 'html', 'lcov'],
      reportsDirectory: './coverage',
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'src/**/*.test.{ts,tsx}',
        'src/**/*.spec.{ts,tsx}',
        'src/test/**',
        'src/types/**',
        'src/main.tsx',
        'src/vite-env.d.ts',
      ],
      // 覆盖率阈值 (可选，根据 Constitution 要求 70%)
      thresholds: {
        lines: 70,
        functions: 70,
        branches: 70,
        statements: 70,
      },
    },

    // 测试超时时间
    testTimeout: 10000,
    hookTimeout: 10000,

    // 报告格式
    reporters: ['default'],

    // 模拟设置
    clearMocks: true,
    restoreMocks: true,
  },
}))
