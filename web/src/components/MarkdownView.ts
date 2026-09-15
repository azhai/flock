import m from 'mithril';
import { renderMarkdown } from '../lib/md';

export function MarkdownView(attrs: any) {
  return m('.markdown-content', m.trust(renderMarkdown(attrs.src)));
}
