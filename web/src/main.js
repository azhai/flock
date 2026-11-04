import Alpine from 'alpinejs';
import './styles/app.css';

import { registerStores } from './core/stores.js';
import { createRouter } from './core/router.js';
import { registerMarkdownDirective } from './core/markdown.js';
import { registerComponents } from './components.js';
import { registerPages } from './pages/index.js';

import sidebarHtml from './views/partials/sidebar.html?raw';
import paginationHtml from './views/partials/pagination.html?raw';
import editorHtml from './views/partials/editor.html?raw';
import tagInputHtml from './views/partials/tag-input.html?raw';

import homeHtml from './views/home.html?raw';
import topicHtml from './views/topic.html?raw';
import newTopicHtml from './views/new-topic.html?raw';
import loginHtml from './views/login.html?raw';
import registerHtml from './views/register.html?raw';
import bookmarksHtml from './views/bookmarks.html?raw';
import profileHtml from './views/profile.html?raw';
import adminHtml from './views/admin.html?raw';

const PARTIALS = {
  sidebar: sidebarHtml,
  pagination: paginationHtml,
  editor: editorHtml,
  'tag-input': tagInputHtml,
};

// 模板片段都是构建期本地字符串（?raw），非用户输入；替换 {{var}} 后注入 DOM。
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

document.getElementById('views').innerHTML = [
  view('home', homeHtml),
  view('topic', topicHtml),
  view('new', newTopicHtml),
  view('login', loginHtml),
  view('register', registerHtml),
  view('profile', profileHtml),
  view('bookmarks', bookmarksHtml),
  view('admin', adminHtml),
].join('\n');

router.start();
Alpine.start();
if (Alpine.store('session').logged) Alpine.store('notifications').start();
