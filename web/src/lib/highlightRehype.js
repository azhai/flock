// 代码高亮 rehype 插件（同步）。高亮器由 highlightProvider 懒加载，
// 未就绪时（首帧）跳过；就绪后同步处理带 language- 前缀的 code 元素。
import { visit } from 'unist-util-visit';
import { toString } from 'hast-util-to-string';
import { getRefractor } from './highlightProvider';

// 常用别名 → 规范名（refractor 不自动处理 aliases，这里手动映射）
const ALIAS = {
  js: 'javascript',
  ts: 'typescript',
  sh: 'bash',
  shell: 'bash',
  py: 'python',
  md: 'markdown',
  yml: 'yaml',
  html: 'markup',
  xml: 'markup',
  'c++': 'cpp',
  'c#': 'csharp',
};

export default function rehypeOnDemandHighlight() {
  return (tree) => {
    const refractor = getRefractor();
    if (!refractor) return; // 高亮器尚未加载（仅首帧），保持原文
    visit(tree, 'element', (node, _index, parent) => {
      if (!parent || node.tagName !== 'code') return;
      const cls = Array.isArray(node.properties?.className) ? node.properties.className : [];
      const langRaw = cls.find((c) => typeof c === 'string' && c.startsWith('language-'));
      if (!langRaw) return;
      let lang = langRaw.slice('language-'.length).toLowerCase();
      lang = ALIAS[lang] || lang;
      if (!refractor.registered(lang)) return; // 未注册语言跳过
      try {
        const root = refractor.highlight(toString(node), lang);
        node.children = root.children;
        node.properties.className = [langRaw];
      } catch {
        /* 高亮失败则保留原文 */
      }
    });
  };
}