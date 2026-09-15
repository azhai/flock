import { defineConfig, loadEnv } from 'vite';
import path from 'path';

// 读取仓库根目录 .env（后端配置），让 dev 代理与后端端口/地址保持一致。
function loadRootEnv(mode) {
  return loadEnv(mode, path.resolve(__dirname, '..'), '');
}

export default defineConfig(({ mode }) => {
  const rootEnv = loadRootEnv(mode);
  const serverPort = rootEnv.SERVER_PORT || '8080';
  const backendBase = rootEnv.BASE_URL || `http://127.0.0.1:${serverPort}`;

  return {
    base: './',
    build: {
      outDir: 'dist',
      emptyOutDir: true,
    },
    server: {
      proxy: {
        '/api/v1': {
          target: backendBase,
          changeOrigin: true,
        },
      },
    },
  };
});
