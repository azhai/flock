import m from 'mithril';
import { api } from '../lib/api';
import { store } from '../lib/store';
import { Layout } from '../components/Layout';
import { Pagination } from '../components/Pagination';
import { Avatar, Tabs, toast } from '../lib/ui';
import { Icon } from '../lib/icon';
import { A } from '../lib/route';
import { renderMarkdown } from '../lib/md';

export const UserProfile: m.Component = {
  oninit(vnode: any) {
    const st = vnode.state as any;
    st.user = null;
    st.tab = 'topics';
    st.items = [];
    st.total = 0;
    st.page = 1;
    st.loading = true;
    st.load = () => {
      const uname = vnode.attrs.username;
      api.getUser(uname).then((u: any) => { st.user = u; m.redraw(); }).catch(() => {});
      st.loading = true;
      const req = st.tab === 'topics' ? api.getUserTopics(uname, st.page) : api.getUserPosts(uname, st.page);
      req.then((d: any) => { st.items = d.data || []; st.total = d.total || 0; st.loading = false; m.redraw(); })
        .catch(() => { st.loading = false; m.redraw(); });
    };
    st.load();
  },
  view(vnode: any) {
    const st = vnode.state;
    const p = m.route.get();
    if (p.indexOf('tab=') >= 0) {
      const tab = new URLSearchParams(p.slice(p.indexOf('?') + 1)).get('tab');
      if (tab && tab !== st.tab) { st.tab = tab; st.page = 1; st.load(); }
    }
    const setTab = (t: string) => { st.tab = t; st.page = 1; m.route.set(`/u/${vnode.attrs.username}?tab=${t}`); };

    const list = st.loading
      ? m('.skeleton', { style: { height: '60px', margin: '12px 0' } })
      : st.items.length === 0
      ? m('.card.card-pad.empty', [m('div', { class: 'big' }, '👤'), m('div', {}, st.tab === 'topics' ? '还没有发布话题' : '还没有回复')])
      : st.items.map((it: any) =>
          st.tab === 'topics'
            ? m('.topic-row', { onclick: () => m.route.set('/topics/' + it.id) }, [
                m('.topic-title', {}, it.title),
                m('.text-xs muted mt-1', `${it.category?.name || ''} · ${it.posts_count || 0} 回复 · ${it.like_count || 0} 赞`),
              ])
            : m('.topic-row', { onclick: () => m.route.set('/topics/' + it.topic_id) }, [
                m('.text-xs muted', `回复于：${it.topic?.title || '话题'}`),
                m('.mt-1', [MarkdownPreview(it.content)]),
              ])
        );

    return m(Layout, [
      m('.container', [
        m('.center-col', { style: { paddingTop: '24px' } }, [
          st.user
            ? m('.card.card-pad.flex.items-center.gap-3', [
                Avatar({ name: st.user.username, size: 'md' }),
                m('div', [
                  m('div', { class: 'fw-700 text-lg' }, st.user.username),
                  m('div', { class: 'text-sm muted' }, [st.user.bio || '这个人很懒，什么都没写', m('span', { class: 'muted ml-2' }, `· 积分 ${st.user.score || 0}`)]),
                ]),
                st.user.id === store.user?.id
                  ? m('div', { class: 'spacer' }, A('/admin', { class: '' }, '') ) // placeholder
                  : null,
              ])
            : m('.skeleton', { style: { height: '70px' } }),
          m(Tabs, { tabs: [{ key: 'topics', label: '话题' }, { key: 'posts', label: '回复' }], active: st.tab, onChange: setTab }),
          m('.card', [m('.card-pad', [list])]),
          m(Pagination, { page: st.page, perPage: 20, total: st.total, onchange: (pg: number) => { st.page = pg; st.load(); } }),
        ]),
      ]),
    ]);
  },
};

function MarkdownPreview(src?: string) {
  return m('div', { class: 'markdown-content', style: { fontSize: '14px' } }, m.trust(renderMarkdown(src)));
}
