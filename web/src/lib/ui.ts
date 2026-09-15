import m from 'mithril';
import { Button as CUButton } from 'construct-ui';

/** 按钮：基于 construct-ui 的 Button 组件，opts = { primary, danger, size:'sm'|'xs', label, onclick, disabled, fluid, class } */
export function Button(opts: any) {
  const intent = opts.primary ? 'primary' : opts.danger ? 'danger' : undefined;
  const size = opts.size === 'sm' || opts.size === 'xs' ? 'sm' : 'md';
  return m(CUButton, {
    intent,
    size,
    label: opts.label,
    fluid: opts.fluid,
    disabled: opts.disabled,
    class: opts.class,
    onclick: opts.onclick,
  });
}

/** 头像：取用户名首字母 */
export function Avatar(attrs: any) {
  const name: string = attrs.name || '?';
  const cls = ['avatar'];
  if (attrs.size === 'md') cls.push('avatar-md');
  if (attrs.size === 'sm') cls.push('avatar-sm');
  return m('div', { class: cls.join(' '), title: name }, (name[0] || '?').toUpperCase());
}

/** 模态框：{ open, title, onClose, wide, body, foot } */
export function Modal(opts: any) {
  if (!opts.open) return null;
  return m(
    '.modal-overlay',
    {
      onclick: (e: any) => {
        if (e.target === e.currentTarget && opts.onClose) opts.onClose();
      },
    },
    [
      m('.modal' + (opts.wide ? '.wide' : ''), [
        m('.modal-head', [
          m('span', {}, opts.title || ''),
          m('.icon-btn', { onclick: opts.onClose }, m(Icon, { name: 'x', size: 18 })),
        ]),
        m('.modal-body', {}, opts.body || null),
        opts.foot ? m('.modal-foot', {}, opts.foot) : null,
      ]),
    ]
  );
}

/** Tab 头：tabs = [{key,label,badge}], active=key, onChange */
export function Tabs(opts: any) {
  return m(
    '.tabs-head',
    opts.tabs.map((t: any) =>
      m(
        'div',
        {
          class: 'tab-item' + (t.key === opts.active ? ' active' : ''),
          onclick: () => opts.onChange(t.key),
        },
        [t.label, t.badge ? m('span', { class: 'badge-dot', style: { position: 'static', marginLeft: '6px' } }, t.badge) : null]
      )
    )
  );
}

// ===== 全局 Toast =====
let items: any[] = [];
let seq = 0;

export function toast(message: string, type = 'info') {
  const t = { id: ++seq, message, type };
  items.push(t);
  m.redraw();
  setTimeout(() => {
    items = items.filter((x) => x.id !== t.id);
    m.redraw();
  }, 2800);
}

export function ToastHost() {
  return m(
    '.toast-wrap',
    items.map((t) => m('.toast.' + t.type, { key: t.id }, t.message))
  );
}

/** 简单确认框（基于 window.confirm 包装，保持 Mithril 流程） */
export function confirm(message: string): boolean {
  return window.confirm(message);
}
