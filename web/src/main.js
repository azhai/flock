import Alpine from 'alpinejs'; // 由 index.html 的 importmap 解析到 /vendor/alpine.module.esm.min.js（ESM，手动 start）

import { registerStores } from './core/stores.js';
import { createRouter } from './core/router.js';
import { registerMarkdownDirective } from './core/markdown.js';
import { registerComponents } from './components.js';
import { registerPages } from './pages/index.js';

// 模板片段与页面模板不再用 Vite 的 ?raw 导入，改为运行时 fetch（同域同源、由 Go embed 提供）。
const PARTIAL_FILES = ['sidebar', 'pagination', 'editor', 'tag-input'];
// 路由名 → 文件名（new 路由对应 new-topic.html）
const VIEW_FILES = {
  home: 'home', topic: 'topic', new: 'new-topic', login: 'login',
  register: 'register', profile: 'profile', bookmarks: 'bookmarks', admin: 'admin',
};

let PARTIALS = {};
let VIEWS = {};

async function loadTemplates() {
  await Promise.all(PARTIAL_FILES.map(async (n) => {
    PARTIALS[n] = await (await fetch(`/src/views/partials/${n}.html`)).text();
  }));
  await Promise.all(Object.entries(VIEW_FILES).map(async ([name, file]) => {
    VIEWS[name] = await (await fetch(`/src/views/${file}.html`)).text();
  }));
}

// 模板片段都是本地字符串（运行时 fetch 得到），非用户输入；替换 {{var}} 后注入 DOM。
function compose(tpl) {
  return tpl.replace(/<!--\s*@include\s+([\w-]+)(?:\s+(\{[^\n]*?\}))?\s*-->/g, (_, name, json) => {
    const fragment = PARTIALS[name] || '';
    if (!json) return fragment;
    let vars;
    try { vars = JSON.parse(json); } catch { return fragment; }
    return Object.entries(vars).reduce(
      (out, [key, value]) => out.split(`{{${key}}}`).join(String(value)),
      fragment,
    );
  });
}

function view(name, tpl) {
  return `<template x-if="$store.route.name === '${name}'">${compose(tpl)}</template>`;
}

window.Alpine = Alpine; // 便于 devtools 调试

registerStores(Alpine);
registerComponents(Alpine);
registerPages(Alpine);
registerMarkdownDirective(Alpine);

const router = createRouter(Alpine);
Alpine.magic('go', () => router.go);
Alpine.magic('router', () => router);

// 等待模板就绪后再挂载视图并启动（top-level await，type=module 支持）。
await loadTemplates();
document.getElementById('views').innerHTML = Object.entries(VIEW_FILES)
  .map(([name]) => view(name, VIEWS[name])).join('\n');

router.start();
Alpine.start();
if (Alpine.store('session').logged) Alpine.store('notifications').start();
