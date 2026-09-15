import m from 'mithril';
import { store } from '../lib/store';
import { notify } from '../lib/notify';
import { A } from '../lib/route';
import { Icon } from '../lib/icon';
import { Avatar, Button } from '../lib/ui';

const NOTIF_META: any = {
  reply: { label: '回复了你的话题', icon: 'messageSquare', color: '#0d8a6f' },
  mention: { label: '提到了你', icon: 'atSign', color: '#d97706' },
  like: { label: '赞了你的回复', icon: 'thumbsUp', color: '#dc2626' },
  follow: { label: '关注了你的话题', icon: 'user', color: '#2563eb' },
  report: { label: '提交了新举报', icon: 'flag', color: '#9333ea' },
};

function timeAgo(dateStr?: string): string {
  if (!dateStr) return '';
  const ms = Date.now() - new Date(dateStr).getTime();
  if (ms < 60000) return '刚刚';
  if (ms < 3600000) return `${Math.floor(ms / 60000)}分钟前`;
  if (ms < 86400000) return `${Math.floor(ms / 3600000)}小时前`;
  if (ms < 604800000) return `${Math.floor(ms / 86400000)}天前`;
  return new Date(dateStr).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
}

export const Header: m.Component = {
  oninit() {
    if (store.user) notify.start();
  },
  view(vnode: any) {
    const st = vnode.state as any;
    const user = store.user;
    const ns = notify.state;

    const submitSearch = () => {
      const kw = (st.q || '').trim();
      m.route.set(kw ? `/?q=${encodeURIComponent(kw)}` : '/');
    };

    const logo = A('/', { class: 'site-logo' }, [m('div', { class: 'mark' }, 'F'), m('span', {}, 'Flock')]);

    const titleBlock = store.headerTitle
      ? [m('.header-title-sep'), m('.header-title', { title: store.headerTitle }, store.headerTitle)]
      : null;

    const search = m('div', { class: 'flex items-center gap-2' }, [
      Icon({ name: 'search', size: 16, class: 'muted' }),
      m('input', {
        class: 'input',
        style: { width: '180px', maxWidth: '180px' },
        placeholder: '搜索',
        value: st.q || '',
        oninput: (e: any) => (st.q = e.target.value),
        onkeydown: (e: any) => e.key === 'Enter' && submitSearch(),
      }),
    ]);

    let right: any;
    if (user) {
      const notifMenu = st.notifOpen
        ? m('.notif-menu', [
            m('.flex', { style: { padding: '10px 12px', borderBottom: '1px solid var(--border)', justifyContent: 'space-between' } }, [
              m('div', { class: 'fw-700 text-sm' }, '通知'),
              ns.unread > 0 ? Button({ size: 'xs', ghost: true, label: '全部已读', onclick: () => notify.markAll() }) : null,
            ]),
            ns.list.length === 0
              ? m('div', { class: 'muted', style: { padding: '32px 12px', textAlign: 'center', fontSize: '13px' } }, '暂无通知')
              : ns.list.slice(0, 20).map((n: any) => {
                  const meta = NOTIF_META[n.type] || { label: '新动态', icon: 'bell', color: 'var(--text2)' };
                  return m(
                    '.notif-item' + (n.read ? '' : '.unread'),
                    {
                      onclick: () => {
                        if (!n.read) notify.markRead(n.id);
                        st.notifOpen = false;
                        if (n.topic_id) m.route.set(`/topics/${n.topic_id}`);
                      },
                    },
                    [
                      Icon({ name: meta.icon, size: 16, style: { color: meta.color, marginTop: '2px' } }),
                      m('div', { style: { minWidth: 0 } }, [
                        m('div', { class: 'text-sm fw-600', style: { display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' } }, n.content || meta.label),
                        m('div', { class: 'text-xs muted' }, `${n.actor?.username || '系统'} ${meta.label} · ${timeAgo(n.created_at)}`),
                      ]),
                    ]
                  );
                }),
          ])
        : null;

      const userMenu = st.userOpen
        ? m('.user-menu', [
            m('button', { onclick: () => { st.userOpen = false; m.route.set(`/u/${user.username}`); } }, [Icon({ name: 'user', size: 16 }), '我的主页']),
            m('button', { onclick: () => { st.userOpen = false; m.route.set('/bookmarks'); } }, [Icon({ name: 'bookmark', size: 16 }), '我的收藏']),
            user.role === 'admin'
              ? m('button', { onclick: () => { st.userOpen = false; m.route.set('/admin'); } }, [Icon({ name: 'shield', size: 16 }), '管理后台'])
              : null,
            m('button', { class: 'danger', onclick: () => { store.logout(); notify.stop(); m.route.set('/'); } }, [Icon({ name: 'logout', size: 16 }), '退出']),
          ])
        : null;

      right = m('.flex.items-center.gap-2', [
        Button({ primary: true, size: 'sm', icon: 'plus', label: '新话题', onclick: () => m.route.set('/new') }),
        m('.menu-wrap', [
          m('.icon-btn', { style: { position: 'relative' }, onclick: () => { st.notifOpen = !st.notifOpen; st.userOpen = false; } }, [
            Icon({ name: 'bell', size: 18 }),
            ns.unread > 0 ? m('.badge-dot', {}, ns.unread > 99 ? '99+' : String(ns.unread)) : null,
          ]),
          notifMenu,
        ]),
        m('.menu-wrap', [
          m('.icon-btn', { onclick: () => { st.userOpen = !st.userOpen; st.notifOpen = false; } }, [Avatar({ name: user.username, size: 'sm' })]),
          userMenu,
        ]),
      ]);
    } else {
      right = m('.flex.gap-1', [
        Button({ ghost: true, size: 'sm', label: '登录', onclick: () => m.route.set('/login') }),
        Button({ primary: true, size: 'sm', label: '注册', onclick: () => m.route.set('/register') }),
      ]);
    }

    return m('header', { class: 'app-header' }, [
      m('.inner', [logo, titleBlock, m('.spacer'), search, right]),
    ]);
  },
};
