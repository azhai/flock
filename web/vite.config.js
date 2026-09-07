import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// 读取仓库根目录的 .env（SITE 后端配置），让 dev 代理与后端端口/地址保持一致。
// loadEnv(mode, envDir, prefix='') —— 空前缀取出全部键。
function loadRootEnv(mode) {
  return loadEnv(mode, path.resolve(__dirname, '..'), '');
}

export default defineConfig(({ mode }) => {
  const rootEnv = loadRootEnv(mode);
  const serverPort = rootEnv.SERVER_PORT || '8080';
  const backendBase = rootEnv.BASE_URL || `http://127.0.0.1:${serverPort}`;

  return {
    plugins: [react()],
    resolve: {
      alias: { '@': path.resolve(__dirname, 'src') },
    },
    server: {
      proxy: {
        '/api/v1': {
          target: backendBase,
          changeOrigin: true,
        },
      },
    },
    build: {
      // 静态打成多个 vendor chunk，使 gzip 尽量不超 100K。
      // 注意：绝不静态匹配 refractor（会让 Rollup 生成非法初始化顺序 → TDZ），
      // 它由 highlightProvider 动态 import，单独成为异步 chunk。
      rollupOptions: {
        output: {
          manualChunks(id) {
            // 语法高亮：仅动态引用 → 独立异步块
            if (id.includes('/src/lib/highlightBundle.js') || id.includes('/refractor/'))
              return 'highlight';
            // react + UI 同一块：react 一旦与 chakra 拆开，二者形成 chunk 级循环，
            // chakra 在 React 就绪前被求值（useLayoutEffect undefined）。故合并。
            if (id.includes('react-dom') || id.includes('react/') || id.includes('/scheduler/')
              || /@chakra-ui|@emotion|framer-motion|react-icons|csstype|@dnd-kit|date-fns|@popperjs|tinycolor|memoize|@babel\/runtime|@ctrl\/tinycolor/.test(id))
              return 'ui';
            // 其余（含整个 md 生态 / unified / parse5 / md-editor）交给 Rollup 默认
            // 自然分组：单块内能对真实 ESM 循环正确排序，避免手工拆块造成的 TDZ。
            return undefined;
          },
        },
      },
    },
  };
});