import m from 'mithril';
import { api } from '../lib/api';
import { store } from '../lib/store';
import { Layout } from '../components/Layout';
import { Sidebar } from '../components/Sidebar';
import { Pagination } from '../components/Pagination';
import { MarkdownView } from '../components/MarkdownView';
import { Avatar, Button, Modal, toast } from '../lib/ui';
import { Icon } from '../lib/icon';
import { A } from '../lib/route';

function timeAgo(dateStr?: string): string {
  if (!dateStr) return '';
  const ms = Date.now() - new Date(dateStr).getTime();
  if (ms < 60000) return '刚刚';
  if (ms < 3600000) return `${Math.floor(ms / 60000)}分钟前`;
  if (ms < 86400000) return `${Math.floor(ms / 3600000)}小时前`;
  return new Date(dateStr).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
}

const ROLE_LABEL: any = { admin: ['管理员', '#9333ea'], moderator: ['管理', '#2563eb'], user: ['会员', 'var(--muted)'] };

const EMOJIS = ['😀', '😂', '😍', '🤔', '👍', '🎉', '🔥', '❤️', '😅', '😎', '🙏', '💡', '✅', '⚠️', '🚀', '💯', '😭', '🤝', '👀', '🌟'];

export const TopicDetail: m.Component = {
  oninit(vnode: any) {
    const st = vnode.state as any;
    st.loading = true;
    st.topic = null;
    st.posts = [];
    st.total = 0;
    st.page = 1;
    st.replyContent = '';
    st.replyPreview = false;
    st.editOpen = false;
    st.editId = null;
    st.editContent = '';
    st.emojiOpen = false;
    st.activePost = null;
    st.reportOpen = false;
    st.reportId = null;
    st.reportReason = '';

    const id = vnode.attrs.id;
    st.load = () => {
      api.getTopic(id).then((t: any) => {
        st.topic = t;
        store.setHeaderTitle(t.title);
        m.redraw();
      }).catch(() => {});
      api.listPosts(id, st.page).then((d: any) => {
        st.posts = d.data || [];
        st.total = d.total || 0;
        st.loading = false;
        m.redraw();
      }).catch(() => { st.loading = false; m.redraw(); });
    };
    st.load();
  },

  view(vnode: any) {
    const st = vnode.state;
    const t = st.topic;
    if (st.loading || !t) {
      return m(Layout, [m('.center-screen', [m('.skeleton', { style: { width: '60%', height: '40px' } })])]);
    }
    const me = store.user;

    const doLike = (post: any) => {
      if (!me) return m.route.set('/login');
      api.toggleLike(post.id).then(() => {
        post.my_like = !post.my_like;
        post.like_count += post.my_like ? 1 : -1;
        m.redraw();
      }).catch((e) => toast(e || '操作失败', 'error'));
    };
    const doVote = (post: any, v: number) => {
      if (!me) return m.route.set('/login');
      const prev = post.my_vote || 0;
      api.votePost(post.id, v).then(() => {
        post.my_vote = prev === v ? 0 : v;
        post.score = (post.score || 0) + (post.my_vote - prev);
        m.redraw();
      }).catch((e) => toast(e || '操作失败', 'error'));
    };
    const doBookmark = () => {
      if (!me) return m.route.set('/login');
      api.bookmarkTopic(t.id).then(() => { t.bookmarked = !t.bookmarked; m.redraw(); toast(t.bookmarked ? '已收藏' : '已取消收藏', 'success'); }).catch((e) => toast(e || '失败', 'error'));
    };
    const doFollow = () => {
      if (!me) return m.route.set('/login');
      api.followTopic(t.id).then(() => { t.followed = !t.followed; m.redraw(); toast(t.followed ? '已关注' : '已取消关注', 'success'); }).catch((e) => toast(e || '失败', 'error'));
    };
    const goPost = (id: number) => {
      st.activePost = id;
      const el = document.getElementById('post-' + id);
      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    };
    const insertEmoji = (emoji: string) => {
      const el = document.getElementById('reply-box') as HTMLTextAreaElement | null;
      if (el) {
        const start = el.selectionStart ?? st.replyContent.length;
        const end = el.selectionEnd ?? st.replyContent.length;
        st.replyContent = st.replyContent.slice(0, start) + emoji + st.replyContent.slice(end);
        m.redraw();
        setTimeout(() => { el.focus(); const pos = start + emoji.length; el.setSelectionRange(pos, pos); }, 0);
      } else {
        st.replyContent += emoji;
        m.redraw();
      }
    };
    const submitReply = () => {
      if (!me) return m.route.set('/login');
      if (!st.replyContent.trim()) return toast('请输入回复内容', 'warning');
      api.createPost(t.id, { content: st.replyContent }).then(() => {
        st.replyContent = '';
        st.page = 1;
        st.load();
        toast('回复成功', 'success');
      }).catch((e) => toast(e || '回复失败', 'error'));
    };
    const openEdit = (post: any) => {
      st.editId = post.id;
      st.editContent = post.content;
      st.editOpen = true;
    };
    const saveEdit = () => {
      if (st.editId == null) return;
      const first = st.posts[0] && st.posts[0].id === st.editId;
      const call = first ? api.updateTopic(t.id, { content: st.editContent }) : api.updatePost(st.editId, { content: st.editContent });
      call.then(() => {
        if (first) { t.first_post = { ...t.first_post, content: st.editContent }; }
        st.posts = st.posts.map((p: any) => (p.id === st.editId ? { ...p, content: st.editContent } : p));
        st.editOpen = false;
        m.redraw();
        toast('已保存', 'success');
      }).catch((e) => toast(e || '保存失败', 'error'));
    };
    const doDelete = (post: any) => {
      const first = st.posts[0] && st.posts[0].id === post.id;
      if (!confirm(first ? '确定删除整个话题？此操作不可撤销。' : '确定删除该回复？')) return;
      const call = first ? api.deleteTopic(t.id) : api.deletePost(post.id);
      call.then(() => {
        toast('已删除', 'success');
        if (first) m.route.set('/');
        else st.load();
      }).catch((e) => toast(e || '删除失败', 'error'));
    };
    const openReport = (post: any) => { st.reportId = post.id; st.reportReason = ''; st.reportOpen = true; };
    const submitReport = () => {
      if (!st.reportReason.trim()) return toast('请填写举报理由', 'warning');
      api.reportPost(st.reportId, st.reportReason).then(() => { st.reportOpen = false; toast('举报已提交', 'success'); }).catch((e) => toast(e || '提交失败', 'error'));
    };

    const isOwner = (post: any) => me && (me.username === post.user?.username || me.role === 'admin' || me.role === 'moderator');

    const postCard = (post: any) => {
      const role = ROLE_LABEL[post.user?.role] || ROLE_LABEL.user;
      return m('.post', { id: 'post-' + post.id }, [
        m('.post-head', [
          Avatar({ name: post.user?.username || '?' }),
          m('div', { style: { minWidth: 0 } }, [
            m('div', { class: 'flex items-center gap-2' }, [
              A('/u/' + (post.user?.username || ''), { class: 'fw-600' }, post.user?.username || '匿名'),
              post.user ? m('span', { class: 'lvl-badge', style: { background: role[1] + '22', color: role[1] } }, role[0]) : null,
              m('span', { class: 'text-xs muted' }, `${timeAgo(post.created_at)} · #${post.post_number}`),
            ]),
          ]),
        ]),
        m('.mt-2', [MarkdownView({ src: post.content })]),
        m('.flex.items-center.gap-3.mt-3', [
          m('.vote-box', [
            m('button', { onclick: () => doVote(post, 1), style: { color: post.my_vote === 1 ? 'var(--accent)' : '' } }, Icon({ name: 'arrowUp', size: 14 })),
            m('span', { class: 'vote-score' }, String(post.score ?? 0)),
            m('button', { onclick: () => doVote(post, -1), style: { color: post.my_vote === -1 ? '#dc2626' : '' } }, Icon({ name: 'arrowDown', size: 14 })),
          ]),
          Button({ size: 'xs', ghost: true, icon: post.my_like ? 'thumbsUp' : 'thumbsUp', label: '赞 ' + (post.like_count || 0), class: post.my_like ? 'btn-primary' : '', onclick: () => doLike(post) }),
          Button({ size: 'xs', ghost: true, icon: 'messageSquare', label: '回复', onclick: () => { const el = document.getElementById('reply-box'); if (el) el.focus(); } }),
          me && me.username !== post.user?.username ? Button({ size: 'xs', ghost: true, icon: 'flag', label: '举报', danger: true, onclick: () => openReport(post) }) : null,
          (isOwner(post) ? [
            Button({ size: 'xs', ghost: true, icon: 'edit', label: '编辑', onclick: () => openEdit(post) }),
            Button({ size: 'xs', danger: true, icon: 'trash', label: '删除', onclick: () => doDelete(post) }),
          ] : null),
        ]),
      ]);
    };

    return m(Layout, [
      m('.layout-grid', [
        m(Sidebar),
        m('.main-col', { style: { padding: '0' } }, [
          m('.detail-grid', [
            // 中栏：帖子区
            m('.detail-main', [
              m('.detail-scroll', [
                m('.card', { style: { borderRadius: '0' } }, [
                  m('.card-pad', [
                    m('.flex.items-center.justify-between.gap-2', [
                      m('h1', { class: 'text-xl fw-700', style: { margin: 0 } }, t.title),
                      m('.flex.gap-1', [
                        t.pinned ? m('span', { class: 'chip' }, [Icon({ name: 'bookmark', size: 12 }), '置顶']) : null,
                        t.closed ? m('span', { class: 'chip chip-muted' }, [Icon({ name: 'lock', size: 12 }), '已关闭']) : null,
                      ]),
                    ]),
                    m('.flex.wrap.items-center.gap-2.mt-2', [
                      t.category ? m('span', { class: 'chip', style: { background: t.category.color + '22', color: t.category.color } }, t.category.name) : null,
                      (t.tags || []).map((tg: any) => m('span', { class: 'text-xs muted' }, '#' + tg.name)),
                    ]),
                    t.first_post ? m('.mt-3', [MarkdownView({ src: t.first_post.content })]) : null,
                  ]),
                ]),
                m('.card', { style: { borderRadius: '0', borderTop: 'none' } }, [
                  st.loading && st.posts.length === 0
                    ? m('.card-pad', [m('.skeleton', { style: { height: '80px' } })])
                    : st.posts.map(postCard),
                ]),
                m(Pagination, { page: st.page, perPage: 20, total: st.total, onchange: (pg: number) => { st.page = pg; st.load(); } }),
              ]),
              // 回复框
              m('.reply-box', [
                m('.flex.items-center.justify-between.mb-2', [
                  m('div', { class: 'fw-600 text-sm' }, '添加回复'),
                  m('.flex.items-center.gap-2', [
                    Button({ size: 'xs', ghost: true, icon: 'smile', label: '表情', class: st.emojiOpen ? 'btn-primary' : '', onclick: () => (st.emojiOpen = !st.emojiOpen) }),
                    Button({ size: 'xs', ghost: true, icon: st.replyPreview ? 'edit' : 'eye', label: st.replyPreview ? '编辑' : '预览', onclick: () => (st.replyPreview = !st.replyPreview) }),
                  ]),
                ]),
                st.emojiOpen
                  ? m('.emoji-bar', EMOJIS.map((e) => m('button', { type: 'button', onclick: () => insertEmoji(e) }, e)))
                  : null,
                st.replyPreview
                  ? m('.md-preview', [MarkdownView({ src: st.replyContent })])
                  : m('textarea', { id: 'reply-box', class: 'textarea', placeholder: '写下你的回复，支持 Markdown…', value: st.replyContent, oninput: (e: any) => (st.replyContent = e.target.value) }),
                m('.btn-row.mt-2', [
                  Button({ primary: true, icon: 'send', label: '回复', onclick: submitReply }),
                ]),
              ]),
            ]),
            // 右栏：操作 + 作者 + 导航
            m('.detail-side', [
              Button({ primary: !t.bookmarked, ghost: t.bookmarked, size: 'sm', icon: 'bookmark', label: t.bookmarked ? '已收藏' : '收藏', class: 'full mb-2', onclick: doBookmark }),
              Button({ ghost: !t.followed, primary: t.followed, size: 'sm', icon: 'eye', label: t.followed ? '已关注' : '关注', class: 'full mb-2', onclick: doFollow }),
              (isOwner(st.posts[0]) ? [
                Button({ size: 'sm', ghost: true, icon: 'edit', label: '编辑话题', class: 'full mb-2', onclick: () => st.posts[0] && openEdit(st.posts[0]) }),
                Button({ size: 'sm', danger: true, icon: 'trash', label: '删除话题', class: 'full mb-2', onclick: () => st.posts[0] && doDelete(st.posts[0]) }),
              ] : null),
              me && me.role === 'admin'
                ? m('div', { class: 'btn-row full mb-2' }, [
                    Button({ size: 'xs', ghost: true, icon: 'star', label: t.pinned ? '取消置顶' : '置顶', class: 'full', onclick: () => { const fn = t.pinned ? api.unpinTopic(t.id) : api.pinTopic(t.id); fn.then(() => { t.pinned = !t.pinned; m.redraw(); }); } }),
                    Button({ size: 'xs', ghost: true, icon: 'lock', label: t.closed ? '开启' : '关闭', class: 'full', onclick: () => { const fn = t.closed ? api.openTopic(t.id) : api.closeTopic(t.id); fn.then(() => { t.closed = !t.closed; m.redraw(); }); } }),
                  ])
                : null,
              m('.divider', { style: { width: '100%', margin: '8px 0' } }),
              m('.text-center', [
                Avatar({ name: t.user?.username || '?', size: 'md' }),
                m('div', { class: 'fw-600 mt-1' }, A('/u/' + (t.user?.username || ''), {}, t.user?.username || '匿名')),
                m('div', { class: 'text-xs muted' }, '作者'),
              ]),
              m('.text-xs muted mt-2', { style: { textAlign: 'center' } }, [
                m('div', {}, `发布于 ${timeAgo(t.created_at)}`),
                m('div', {}, `${t.views_count || 0} 浏览 · ${t.posts_count || 0} 回复`),
              ]),
              m('.divider', { style: { width: '100%', margin: '8px 0' } }),
              m('.nav-label', '楼层导航'),
              m('.nav-dots', st.posts.map((p: any) => m('.nav-dot', {
                class: st.activePost === p.id ? 'active' : '',
                'data-n': '#' + p.post_number,
                title: '第 ' + p.post_number + ' 楼',
                onclick: () => goPost(p.id),
              }))),
            ]),
          ]),
        ]),
      ]),
      // 编辑弹窗
      Modal({
        open: st.editOpen,
        title: '编辑内容',
        onClose: () => (st.editOpen = false),
        body: [
          m('textarea', { class: 'textarea', style: { minHeight: '200px' }, value: st.editContent, oninput: (e: any) => (st.editContent = e.target.value) }),
        ],
        foot: [
          Button({ ghost: true, label: '取消', onclick: () => (st.editOpen = false) }),
          Button({ primary: true, label: '保存', onclick: saveEdit }),
        ],
      }),
      // 举报弹窗
      Modal({
        open: st.reportOpen,
        title: '举报内容',
        onClose: () => (st.reportOpen = false),
        body: [
          m('label', { class: 'field-label' }, '举报理由'),
          m('textarea', { class: 'textarea', placeholder: '请说明举报原因…', value: st.reportReason, oninput: (e: any) => (st.reportReason = e.target.value) }),
        ],
        foot: [
          Button({ ghost: true, label: '取消', onclick: () => (st.reportOpen = false) }),
          Button({ danger: true, label: '提交举报', onclick: submitReport }),
        ],
      }),
    ]);
  },
};
