import { marked } from 'marked';
import DOMPurify from 'dompurify';
import microlight from 'microlight';

marked.setOptions({ gfm: true, breaks: true });

/** 用户内容唯一的渲染入口：解析 → 消毒。绝不把结果之外的字符串塞进 x-html。 */
export function renderMarkdown(raw) {
  const html = marked.parse(String(raw ?? ''));
  return DOMPurify.sanitize(html, { USE_PROFILES: { html: true } });
}

/** 消毒之后再用 microlight 高亮（它只读 textContent、只写 span，不解析 HTML），并收紧外链。 */
export function enhanceMarkdown(root) {
  root.querySelectorAll('pre code').forEach((code) => code.classList.add('microlight'));
  const codes = [...document.querySelectorAll('pre code.microlight')];
  const snapshots = codes.map((code) => code.textContent);
  try {
    microlight.reset();
  } catch {
    // reset 会先清空 innerHTML；异常时恢复纯文本，避免代码块变空。
    codes.forEach((code, i) => { code.textContent = snapshots[i]; });
  }
  root.querySelectorAll('a[href]').forEach((a) => {
    const href = a.getAttribute('href') || '';
    if (/^https?:\/\//i.test(href)) {
      a.setAttribute('target', '_blank');
      a.setAttribute('rel', 'noopener noreferrer');
    }
  });
}

/** x-markdown="expr" —— 响应式的安全 Markdown 渲染指令。 */
export function registerMarkdownDirective(Alpine) {
  Alpine.directive('markdown', (el, { expression }, { evaluateLater, effect }) => {
    const getValue = evaluateLater(expression);
    effect(() => {
      getValue((raw) => {
        el.innerHTML = renderMarkdown(raw);
        enhanceMarkdown(el);
      });
    });
  });
}
