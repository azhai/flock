import { useEffect, useReducer } from 'react';
import type { ReactNode } from 'react';
import { getRefractor, ensureSyntaxHighlight } from '../lib/highlightProvider';

// 常用别名 → refractor 规范名（react-markdown 只给 language-<name>，直接透传；这里补齐常见别名）
const ALIAS: Record<string, string> = {
  js: 'javascript',
  ts: 'typescript',
  sh: 'bash',
  shell: 'bash',
  py: 'python',
  md: 'markdown',
  yml: 'yaml',
  html: 'markup',
  xml: 'markup',
  golang: 'go',
  'c++': 'cpp',
  'c#': 'csharp',
};

// 把 refractor 输出的 hast 子节点递归渲染成 React 节点（span+text）
function renderHast(nodes: any[], prefix: string): ReactNode {
  let k = 0;
  return nodes.map((n: any) => {
    if (n.type === 'text') return n.value;
    const cls = Array.isArray(n.properties?.className) ? n.properties.className.join(' ') : '';
    const key = `${prefix}-${k++}`;
    return (
      <span key={key} className={cls}>
        {Array.isArray(n.children) ? renderHast(n.children, key) : null}
      </span>
    );
  });
}

/**
 * react-markdown 的 `code` 组件覆盖：块级代码（带 language-xxx）用 refractor 高亮。
 * 高亮器由 highlightProvider 懒加载；未就绪时先渲染纯文本，加载完成后强制重渲染高亮。
 */
export default function CodeBlock({
  node: _node,
  inline,
  className,
  children,
  ...props
}: any) {
  const [, force] = useReducer((x: number) => x + 1, 0);

  useEffect(() => {
    ensureSyntaxHighlight().then(force).catch(() => {});
  }, []);

  const match = /language-(\w+)/.exec(className || '');
  const code = String(children ?? '');

  if (!inline && match) {
    const raw = match[1].toLowerCase();
    const lang = ALIAS[raw] || raw;
    const refractor = getRefractor();
    if (refractor && refractor.registered(lang)) {
      try {
        // markdown 代码块末尾自带一个换行，去掉避免高亮结果出现多余空行
        const tree: any = refractor.highlight(code.replace(/\n$/, ''), lang);
        return (
          <code className={className} {...props}>
            {renderHast(tree.children, 'tk')}
          </code>
        );
      } catch {
        /* 高亮失败则回退纯文本 */
      }
    }
  }

  return (
    <code className={className} {...props}>
      {children}
    </code>
  );
}