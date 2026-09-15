import { marked } from 'marked';
import hljs from 'highlight.js';

marked.setOptions({ breaks: true, gfm: true });

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

const renderer = new marked.Renderer();
// @ts-ignore - marked v12 签名兼容
renderer.code = (code: string, infostring: string | undefined) => {
  const lang = (infostring || '').trim().split(/\s+/)[0];
  let html: string;
  if (lang && hljs.getLanguage(lang)) {
    try {
      html = hljs.highlight(code, { language: lang }).value;
    } catch {
      html = escapeHtml(code);
    }
  } else {
    try {
      html = hljs.highlightAuto(code).value;
    } catch {
      html = escapeHtml(code);
    }
  }
  const cls = lang ? `language-${lang}` : '';
  return `<pre><code class="hljs ${cls}">${html}</code></pre>`;
};
marked.use({ renderer });

export function renderMarkdown(src?: string): string {
  if (!src) return '';
  return (marked.parse(src) as string) || '';
}
