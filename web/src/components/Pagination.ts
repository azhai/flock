import m from 'mithril';
import { Icon } from '../lib/icon';

export const Pagination: m.Component = {
  view(vnode: any) {
    const { page, perPage, total, onchange } = vnode.attrs;
    const totalPages = Math.max(1, Math.ceil(total / perPage));
    if (totalPages <= 1) return null;

    const pages: any[] = [1];
    const left = Math.max(2, page - 2);
    const right = Math.min(totalPages - 1, page + 2);
    if (left > 2) pages.push('...');
    for (let p = left; p <= right; p++) pages.push(p);
    if (right < totalPages - 1) pages.push('...');
    if (totalPages > 1) pages.push(totalPages);

    return m('.pagination', [
      m('div', { class: 'text-sm muted' }, `共 ${total} 条 · 第 ${page}/${totalPages} 页`),
      m('.flex.items-center.gap-1', [
        m('button', { class: 'page-btn', disabled: page <= 1, onclick: () => page > 1 && onchange(page - 1) }, Icon({ name: 'chevronLeft', size: 16 })),
        pages.map((p) =>
          p === '...'
            ? m('span', { class: 'page-btn muted' }, '...')
            : m('button', { class: 'page-btn' + (p === page ? ' active' : ''), onclick: () => onchange(p) }, String(p))
        ),
        m('button', { class: 'page-btn', disabled: page >= totalPages, onclick: () => page < totalPages && onchange(page + 1) }, Icon({ name: 'chevronRight', size: 16 })),
      ]),
    ]);
  },
};
