import m from 'mithril';
import { api } from '../lib/api';
import { Icon } from '../lib/icon';

export const TagInput: m.Component = {
  oninit(vnode: any) {
    const st = vnode.state as any;
    st.suggestions = [];
    st.query = '';
    st.open = false;
  },
  view(vnode: any) {
    const { value = [], onchange, placeholder } = vnode.attrs;
    const st = vnode.state;

    const add = (name: string) => {
      const slug = name.trim().toLowerCase();
      if (!slug || value.includes(slug)) return;
      onchange([...value, slug]);
      st.query = '';
      st.open = false;
    };
    const remove = (slug: string) => onchange(value.filter((x: string) => x !== slug));
    const loadSug = () =>
      api.listTags().then((t: any) => {
        st.suggestions = Array.isArray(t) ? t : [];
        m.redraw();
      }).catch(() => {});

    const filtered = st.open
      ? st.suggestions.filter((t: any) => t.name && t.name.toLowerCase().includes(st.query.toLowerCase()) && !value.includes(t.slug)).slice(0, 8)
      : [];

    return m('div', { style: { position: 'relative' } }, [
      m(
        '.flex.wrap.gap-1',
        value.map((slug: string) =>
          m('.tag-pill', { style: { background: 'var(--accent-soft)', color: 'var(--accent)' } }, [
            '#' + slug,
            m('span', { style: { cursor: 'pointer' }, onclick: () => remove(slug) }, Icon({ name: 'x', size: 12 })),
          ])
        )
      ),
      m('input', {
        class: 'input',
        style: { marginTop: value.length ? '6px' : '0' },
        placeholder: placeholder || '添加标签，回车确认',
        value: st.query,
        onfocus: () => {
          st.open = true;
          if (!st.suggestions.length) loadSug();
        },
        oninput: (e: any) => {
          st.query = e.target.value;
          st.open = true;
        },
        onkeydown: (e: any) => {
          if (e.key === 'Enter') {
            e.preventDefault();
            add(st.query);
          } else if (e.key === 'Backspace' && !st.query && value.length) {
            remove(value[value.length - 1]);
          }
        },
        onblur: () => setTimeout(() => {
          st.open = false;
          m.redraw();
        }, 150),
      }),
      st.open && filtered.length
        ? m('.notif-menu', { style: { position: 'absolute', top: '100%', width: '100%', right: 'auto', left: '0', zIndex: 160, marginTop: '4px' } },
            filtered.map((t: any) =>
              m('.notif-item', { onmousedown: (e: any) => { e.preventDefault(); add(t.name); } }, [
                m('span', { class: 'fw-600' }, '#' + t.name),
                m('span', { class: 'muted text-xs' }, t.topics_count + ' 话题'),
              ])
            )
          )
        : null,
    ]);
  },
};
