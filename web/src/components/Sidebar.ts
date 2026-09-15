import m from 'mithril';
import { api } from '../lib/api';
import { Icon } from '../lib/icon';

export const Sidebar: m.Component = {
  oninit(vnode: any) {
    const st = vnode.state as any;
    st.cats = [];
    st.tags = [];
    st.loaded = false;
    Promise.all([api.listCategories(), api.listTags()])
      .then(([cats, tags]: any) => {
        st.cats = cats || [];
        st.tags = Array.isArray(tags) ? tags : [];
        st.loaded = true;
        m.redraw();
      })
      .catch(() => {});
  },
  view(vnode: any) {
    const st = vnode.state;
    const route = m.route.get();
    const qi = route.indexOf('?');
    const params = new URLSearchParams(qi >= 0 ? route.slice(qi + 1) : '');
    const activeCat = params.get('cat');
    const activeTag = params.get('tag');

    const go = (extra: any) => {
      const sp = new URLSearchParams();
      if (extra.cat) sp.set('cat', extra.cat);
      if (extra.tag) sp.set('tag', extra.tag);
      m.route.set(sp.toString() ? `/?${sp.toString()}` : '/');
    };

    const Item = (label: string, dot: any, active: boolean, onclick: any) =>
      m(
        'button',
        { class: 'side-item' + (active ? ' active' : ''), onclick },
        [
          dot ? m('span', { class: 'cat-dot', style: { background: dot } }) : Icon({ name: 'messageSquare', size: 16 }),
          m('span', { class: 'truncate' }, label),
        ]
      );

    return m('div', { class: 'sidebar-col' }, [
      m('.side-section-title', {}, '分类'),
      m('div', { style: { padding: '0 8px' } }, [
        Item('全部', null, !activeCat && !activeTag, () => go({})),
        st.cats.map((c: any) => Item(c.name, c.color, activeCat === c.slug, () => go({ cat: c.slug }))),
      ]),
      st.tags.length > 0
        ? [
            m('.side-section-title', {}, '热门标签'),
            m(
              '.tag-cloud',
              st.tags.slice(0, 12).map((tg: any) => {
                const active = activeTag === tg.slug;
                return m(
                  'span',
                  { class: 'tag-pill' + (active ? ' active' : ''), onclick: () => (active ? go({}) : go({ tag: tg.slug })) },
                  ['#' + tg.name, m('span', { class: 'muted', style: { fontWeight: 400 } }, '· ' + tg.topics_count)]
                );
              })
            ),
          ]
        : null,
    ]);
  },
};
