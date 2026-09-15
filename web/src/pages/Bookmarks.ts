import m from 'mithril';
import { api } from '../lib/api';
import { store } from '../lib/store';
import { Layout } from '../components/Layout';
import { Pagination } from '../components/Pagination';
import { Icon } from '../lib/icon';

export const Bookmarks: m.Component = {
  oninit(vnode: any) {
    const st = vnode.state as any;
    st.loading = true;
    st.topics = [];
    st.total = 0;
    st.page = 1;
    st.load = () => {
      st.loading = true;
      api.listBookmarks()
        .then((d: any) => { st.topics = d.data || []; st.total = d.total || 0; st.loading = false; m.redraw(); })
        .catch(() => { st.loading = false; m.redraw(); });
    };
    st.load();
  },
  view(vnode: any) {
    const st = vnode.state;
    if (!store.user) { m.route.set('/login'); return m('div'); }
    const list = st.loading
      ? [1, 2].map(() => m('.skeleton', { style: { height: '60px', margin: '12px 0' } }))
      : st.topics.length === 0
      ? m('.card.card-pad.empty', [m('div', { class: 'big' }, '🔖'), m('div', {}, '还没有收藏任何话题')])
      : st.topics.map((t: any) =>
          m('.topic-row', { onclick: () => m.route.set('/topics/' + t.id) }, [
            m('.topic-grid', [
              m('div', { style: { minWidth: 0 } }, [
                m('.topic-title', {}, t.title),
                m('.text-xs muted mt-1', `${t.user?.username || ''} · ${t.posts_count || 0} 回复`),
              ]),
              m('.topic-stat', [Icon({ name: 'thumbsUp', size: 14 }), String(t.like_count || 0)]),
              m('.topic-stat', [Icon({ name: 'messageSquare', size: 14 }), String(t.posts_count || 0)]),
              m('.topic-stat', [Icon({ name: 'eye', size: 14 }), String(t.views_count || 0)]),
            ]),
          ])
        );
    return m(Layout, [
      m('.container', [
        m('.center-col', { style: { paddingTop: '24px' } }, [
          m('h1', { class: 'text-xl fw-700 mb-3' }, '我的收藏'),
          m('.card', [m('.card-pad', [list])]),
          m(Pagination, { page: st.page, perPage: 20, total: st.total, onchange: (pg: number) => { st.page = pg; st.load(); } }),
        ]),
      ]),
    ]);
  },
};
