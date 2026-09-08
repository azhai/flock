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
      // 静态打成多个 vendor chunk，使每个 gzip 后 ≤100K，总数 ≤10。
      // 注意：绝不静态匹配 refractor（会让 Rollup 生成非法初始化顺序 → TDZ），
      // 它由 highlightProvider 动态 import，单独成为异步 chunk。
      rollupOptions: {
        output: {
          manualChunks(id) {
            // 语法高亮：仅动态引用 → 独立异步块
            if (id.includes('/src/lib/highlightBundle.js') || id.includes('/refractor/'))
              return 'highlight';
            // React 运行时（含 router / zustand 状态）
            if (/node_modules\/(react|react-dom|scheduler|zustand|react-router|react-router-dom|@remix-run)\//.test(id))
              return 'react-core';
            // Chakra UI / emotion / motion：CSS-in-JS 体系，与 react-core 分离已验证无循环 chunk
            if (id.includes('@chakra-ui')) return 'chakra';
            if (id.includes('/@emotion/') || id.includes('/framer-motion/') || id.includes('/csstype/'))
              return 'emotion';
            if (id.includes('/react-icons/')) return 'icons';
            // Markdown 编辑器（nohighlight 子路径，不含 prism）
            if (id.includes('/@uiw/')) return 'md-editor';
            // Markdown 渲染/解析生态（含 unified / micromark / parse5 / hast 等）
            if (/node_modules\/(react-markdown|remark-|rehype-|unified|micromark|mdast-|unist-|hast-|estree-|parse5|vfile|bail|trough)\//.test(id))
              return 'md-render';
            // 其余（含应用代码）交给 Rollup 默认分组
            return undefined;
          },
        },
      },
    },
  };
});