import m from 'mithril';
import { api } from '../lib/api';
import { store } from '../lib/store';
import { Layout } from '../components/Layout';
import { Sidebar } from '../components/Sidebar';
import { Pagination } from '../components/Pagination';
import { Icon } from '../lib/icon';
import { Avatar, Button } from '../lib/ui';
import { A } from '../lib/route';

function getParams() {
  const r = m.route.get();
  const qi = r.indexOf('?');
  const p = new URLSearchParams(qi >= 0 ? r.slice(qi + 1) : '');
  return {
    cat: p.get('cat'),
    q: p.get('q'),
    tag: p.get('tag'),
    sort: p.get('sort') || 'created_desc',
    page: parseInt(p.get('page') || '1', 10),
  };
}

const SORTS: any[] = [
  { key: 'created_desc', label: '最新' },
  { key: 'last_activity_desc', label: '活跃' },
  { key: 'like_count', label: '热门' },
  { key: 'posts_count', label: '获赞' },
];

export const Home: m.Component = {
  oninit(vnode: any) {
    const st = vnode.state as any;
    st.loading = true;
    st.topics = [];
    st.total = 0;
    st.last = '';
    st.load = () => {
      const p = getParams();
      st.loading = true;
      m.redraw();
      api
        .listTopics({ category_id: p.cat, q: p.q, tag: p.tag, sort: p.sort, page: p.page })
        .then((d: any) => {
          st.topics = d.data || [];
          st.total = d.total || 0;
          st.loading = false;
          m.redraw();
        })
        .catch(() => {
          st.loading = false;
          m.redraw();
        });
    };
    st.load();
  },
  view(vnode: any) {
    const st = vnode.state;
    const p = getParams();
    const key = `${p.cat}|${p.q}|${p.tag}|${p.sort}|${p.page}`;
    if (key !== st.last) {
      st.last = key;
      st.load();
    }

    const setParams = (override: any) => {
      const base = getParams();
      const sp = new URLSearchParams();
      const cat = override.cat !== undefined ? override.cat : base.cat;
      const q = override.q !== undefined ? override.q : base.q;
      const tag = override.tag !== undefined ? override.tag : base.tag;
      const sort = override.sort !== undefined ? override.sort : base.sort;
      const page = override.page !== undefined ? override.page : 1;
      if (cat) sp.set('cat', cat);
      if (q) sp.set('q', q);
      if (tag) sp.set('tag', tag);
      if (sort && sort !== 'created_desc') sp.set('sort', sort);
      if (page && page > 1) sp.set('page', String(page));
      m.route.set(sp.toString() ? '/?' + sp.toString() : '/');
    };

    const heading = p.q
      ? `搜索：“${p.q}”`
      : p.tag
      ? `#${p.tag}`
      : p.cat
      ? '分类话题'
      : '最新讨论';

    const list = st.loading
      ? [1, 2, 3].map(() => m('.topic-row', [m('.skeleton', { style: { height: '16px', width: '40%', marginBottom: '8px' } }), m('.skeleton', { style: { height: '12px', width: '70%' } })]))
      : st.topics.length === 0
      ? m('.card.card-pad.empty', [m('div', { class: 'big' }, '🌿'), m('div', {}, '这里还没有话题'), store.user ? Button({ primary: true, size: 'sm', icon: 'plus', label: '发布第一个话题', onclick: () => m.route.set('/new'), class: 'mt-3' }) : null])
      : st.topics.map((t: any) =>
          m(
            '.topic-row',
            { onclick: () => m.route.set('/topics/' + t.id) },
            [
              m('.topic-grid', [
                m('div', { style: { minWidth: 0 } }, [
                  m('.topic-title', {}, t.title),
                  m('.flex.items-center.gap-2.mt-1', [
                    t.category ? m('span', { class: 'chip', style: { background: t.category.color + '22', color: t.category.color } }, t.category.name) : null,
                    m('span', { class: 'text-xs muted' }, `${t.user?.username || ''} · ${t.created_at ? new Date(t.created_at).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }) : ''}`),
                  ]),
                  t.tags && t.tags.length
                    ? m('.flex.wrap.gap-1.mt-1', t.tags.map((tg: any) => m('span', { class: 'text-xs muted' }, '#' + tg.name)))
                    : null,
                ]),
                m('.topic-stat', [Icon({ name: 'thumbsUp', size: 14 }), String(t.like_count || 0)]),
                m('.topic-stat', [Icon({ name: 'messageSquare', size: 14 }), String(t.posts_count || 0)]),
                m('.topic-stat', [Icon({ name: 'eye', size: 14 }), String(t.views_count || 0)]),
              ]),
            ]
          )
        );

    return m(Layout, [
      m('.layout-grid', [
        m(Sidebar),
        m('.main-col', [
          m('.flex.items-center.justify-between.mb-3', [
            m('h2', { class: 'text-xl fw-700', style: { margin: 0 } }, heading),
            store.user ? Button({ primary: true, size: 'sm', icon: 'plus', label: '新话题', onclick: () => m.route.set('/new') }) : null,
          ]),
          m('.sort-tabs.mb-2', SORTS.map((s) => m('div', { class: 'sort-tab' + (s.key === p.sort ? ' active' : ''), onclick: () => setParams({ sort: s.key, page: 1 }) }, s.label))),
          m('.card', list),
          m(Pagination, { page: p.page, perPage: 20, total: st.total, onchange: (pg: number) => setParams({ page: pg }) }),
        ]),
      ]),
    ]);
  },
};
