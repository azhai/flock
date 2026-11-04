import { defineConfig, loadEnv } from 'vite';
import path from 'path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));

// 读取仓库根目录的 .env，让 dev 代理与后端端口/地址保持一致。
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, path.resolve(here, '..'), '');
  const serverPort = env.SERVER_PORT || '8080';
  const backendBase = env.BASE_URL || `http://127.0.0.1:${serverPort}`;

  return {
    server: {
      proxy: {
        '/api/v1': { target: backendBase, changeOrigin: true },
      },
    },
    build: {
      outDir: 'dist',
      emptyOutDir: true,
    },
  };
});
