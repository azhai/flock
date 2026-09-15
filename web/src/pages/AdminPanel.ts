import m from 'mithril';
import { api } from '../lib/api';
import { store } from '../lib/store';
import { Layout } from '../components/Layout';
import { Pagination } from '../components/Pagination';
import { Button, Tabs, toast } from '../lib/ui';
import { Icon } from '../lib/icon';

export const AdminPanel: m.Component = {
  oninit(vnode: any) {
    const st = vnode.state as any;
    st.tab = 'overview';
    st.users = [];
    st.userTotal = 0;
    st.userPage = 1;
    st.reports = [];
    st.reportStatus = 'open';
    st.cats = [];
    st.newCatName = '';
    st.newCatColor = '#0d8a6f';
    st.newCatSlug = '';
    st.stats = null;
    st.loadUsers = () => {
      api.listUsers({ page: st.userPage }).then((d: any) => { st.users = d.data || []; st.userTotal = d.total || 0; m.redraw(); }).catch(() => {});
    };
    st.loadReports = () => {
      api.listReports(st.reportStatus).then((d: any) => { st.reports = Array.isArray(d) ? d : (d.data || []); m.redraw(); }).catch(() => {});
    };
    st.loadCats = () => {
      api.listCategories().then((c: any) => { st.cats = c || []; m.redraw(); }).catch(() => {});
    };
    st.loadUsers();
    st.loadCats();
  },
  view(vnode: any) {
    const st = vnode.state;
    if (store.user?.role !== 'admin') { m.route.set('/'); return m('div'); }

    const setTab = (t: string) => {
      st.tab = t;
      if (t === 'reports') st.loadReports();
      m.redraw();
    };

    let body: any;
    if (st.tab === 'overview') {
      body = m('.flex.wrap.gap-3', [
        statCard('用户总数', st.users.length ? st.userTotal : '—', '/admin?tab=users'),
        statCard('举报待处理', st.reports.filter((r: any) => r.status === 'open').length || '—', '/admin?tab=reports'),
        statCard('分类数', st.cats.length, '/admin?tab=categories'),
      ]);
    } else if (st.tab === 'users') {
      body = m('.card', [
        m('.card-pad', st.users.length === 0 ? m('.empty', '暂无用户') :
          st.users.map((u: any) =>
            m('.topic-row', { style: { display: 'flex', alignItems: 'center', gap: '12px' } }, [
              m('div', { class: 'avatar' }, (u.username[0] || '?').toUpperCase()),
              m('div', { style: { flex: 1, minWidth: 0 } }, [
                m('div', { class: 'fw-600' }, u.username),
                m('div', { class: 'text-xs muted' }, `积分 ${u.score || 0} · ${u.disabled ? '已禁用' : '正常'}`),
              ]),
              m('.chip' + (u.role === 'admin' ? '' : ' chip-muted'), u.role),
              Button({ size: 'xs', ghost: true, label: u.disabled ? '启用' : '禁用', onclick: () => api.adminUpdateUser(u.username, { disabled: !u.disabled }).then(() => { toast('已更新', 'success'); st.loadUsers(); }).catch((e) => toast(e || '失败', 'error')) }),
              m('select', { class: 'select', style: { width: 'auto' }, value: u.role, oninput: (e: any) => api.adminUpdateUser(u.username, { role: e.target.value }).then(() => toast('角色已更新', 'success')).catch((e) => toast(e || '失败', 'error')) },
                [m('option', { value: 'user' }, 'user'), m('option', { value: 'moderator' }, 'moderator'), m('option', { value: 'admin' }, 'admin')]
              ),
            ])
          )
        ),
      ]);
    } else if (st.tab === 'reports') {
      body = m('div', [
        m('.btn-row.mb-3', [
          Button({ size: 'sm', ghost: st.reportStatus === 'open', label: '待处理', onclick: () => { st.reportStatus = 'open'; st.loadReports(); } }),
          Button({ size: 'sm', ghost: st.reportStatus === 'resolved', label: '已处理', onclick: () => { st.reportStatus = 'resolved'; st.loadReports(); } }),
          Button({ size: 'sm', ghost: st.reportStatus === 'dismissed', label: '已驳回', onclick: () => { st.reportStatus = 'dismissed'; st.loadReports(); } }),
        ]),
        m('.card', [m('.card-pad', st.reports.length === 0 ? m('.empty', '暂无举报') :
          st.reports.map((r: any) =>
            m('.topic-row', [
              m('.flex.items-center.justify-between', [
                m('div', [m('div', { class: 'fw-600' }, r.reason), m('div', { class: 'text-xs muted' }, `举报人：${r.reporter?.username || '—'} · 状态：${r.status}`) ]),
                m('.btn-row', [
                  Button({ size: 'xs', ghost: true, label: '查看帖子', onclick: () => r.post_id && m.route.set('/topics/' + r.topic_id) }),
                  Button({ size: 'xs', primary: true, label: '处理', onclick: () => api.resolveReport(r.id).then(() => { toast('已处理', 'success'); st.loadReports(); }).catch((e) => toast(e || '失败', 'error')) }),
                  Button({ size: 'xs', danger: true, label: '驳回', onclick: () => api.dismissReport(r.id).then(() => { toast('已驳回', 'success'); st.loadReports(); }).catch((e) => toast(e || '失败', 'error')) }),
                ]),
              ]),
            ])
          )
        )]),
      ]);
    } else {
      body = m('.card.card-pad', [
        m('h3', { class: 'mt-0' }, '分类管理'),
        m('.flex.gap-2.mb-3', [
          m('input', { class: 'input', style: { maxWidth: '160px' }, placeholder: '名称', value: st.newCatName, oninput: (e: any) => (st.newCatName = e.target.value) }),
          m('input', { class: 'input', style: { maxWidth: '140px' }, placeholder: 'slug', value: st.newCatSlug, oninput: (e: any) => (st.newCatSlug = e.target.value) }),
          m('input', { class: 'input', type: 'color', style: { maxWidth: '48px', padding: '2px' }, value: st.newCatColor, oninput: (e: any) => (st.newCatColor = e.target.value) }),
          Button({ size: 'sm', primary: true, label: '新增', onclick: () => {
            if (!st.newCatName.trim()) return toast('请填写名称', 'warning');
            api.createCategory({ name: st.newCatName.trim(), slug: st.newCatSlug.trim() || st.newCatName.trim().toLowerCase(), color: st.newCatColor })
              .then(() => { toast('已添加', 'success'); st.newCatName = ''; st.newCatSlug = ''; st.loadCats(); }).catch((e) => toast(e || '失败', 'error'));
          } }),
        ]),
        m('.flex.wrap.gap-2', st.cats.map((c: any) => m('span', { class: 'tag-pill', style: { background: c.color + '22', color: c.color } }, [m('span', { class: 'cat-dot', style: { background: c.color } }), c.name]))),
      ]);
    }

    return m(Layout, [
      m('.container', [
        m('.center-col', { style: { paddingTop: '24px' } }, [
          m('h1', { class: 'text-xl fw-700 mb-3' }, '管理后台'),
          Tabs({ tabs: [{ key: 'overview', label: '概览' }, { key: 'users', label: '用户' }, { key: 'reports', label: '举报' }, { key: 'categories', label: '分类' }], active: st.tab, onChange: setTab }),
          body,
        ]),
      ]),
    ]);
  },
};

function statCard(label: string, num: any, href: string) {
  return m('div', { class: 'stat-card' }, [
    m('div', { class: 'num' }, String(num)),
    m('div', { class: 'lbl' }, label),
  ]);
}
