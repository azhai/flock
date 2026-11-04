import { api } from '../core/api.js';
import {
  editorMixin, EMOJIS, badgeStyle, initial, formatDate, formatDateTime,
  levelOf, safeColor,
} from '../core/utils.js';

const SNAP = 64;      // 距刻度多少像素内吸附
const RELEASE = 160;  // 离开当前吸附刻度多远才允许吸附下一个

export function topic() {
  return {
    ...editorMixin,
    emojis: EMOJIS,
    topic: null,
    posts: [],
    cats: [],
    loading: true,
    reply: '',
    submitting: false,
    editOpen: false,
    eTitle: '',
    eCat: '',
    eTags: [],
    reportOpen: false,
    reportTarget: null,
    reportReason: '',
    nodeTops: [],
    activePostId: null,
    prog: 0,
    railH: 200,
    dragging: false,
    snap: null,
    _onResize: null,

    get id() { return this.$store.route.params.id; },
    get canEditTopic() {
      const u = this.$store.session.user;
      return !!u && (this.$store.session.isStaff || u.username === this.topic?.user?.username);
    },
    get thumbHeight() {
      const el = this.$refs.scroller;
      const scH = el?.scrollHeight || 1;
      const cliH = el?.clientHeight || scH;
      return Math.max(this.railH * (cliH / scH), 18);
    },
    get thumbTop() { return this.prog * Math.max(this.railH - this.thumbHeight, 0); },
    get activeLabel() {
      const p = this.posts.find((x) => x.id === this.activePostId);
      return p ? (p.post_number === 1 ? '主贴' : '回复 #' + p.post_number) : '';
    },

    init() {
      this.$watch('$store.route.version', () => this.load());
      this._onResize = () => this.calibrate();
      window.addEventListener('resize', this._onResize);
      api.listCategories()
        .then((d) => { this.cats = Array.isArray(d) ? d : []; })
        .catch(() => {});
      this.load();
    },
    destroy() {
      if (this._onResize) window.removeEventListener('resize', this._onResize);
      this.$store.ui.headerTitle = null;
    },

    async load() {
      const id = this.id;
      if (!id) return;
      this.loading = true;
      this.$store.ui.headerTitle = null;
      try { this.topic = await api.getTopic(id); }
      catch { this.topic = null; this.$store.ui.fail('话题不存在'); }
      try {
        const d = await api.listPosts(id);
        this.posts = d?.data || [];
      } catch { this.posts = []; }
      this.loading = false;
      this.$nextTick(() => {
        if (this.$refs.scroller) this.$refs.scroller.scrollTop = 0;
        this.prog = 0;
        this.activePostId = this.posts[0]?.id || null;
        this.calibrate();
        setTimeout(() => this.calibrate(), 300); // 等排版/图片稳定后校准
      });
    },

    // ── 钓鱼线刻度条 ──────────────────────────────────
    nodeEl(id) {
      return this.$refs.scroller?.querySelector(`[data-post-id="${id}"]`) || null;
    },
    calibrate() {
      const el = this.$refs.scroller;
      const rail = this.$refs.rail;
      if (!el || !rail) return;
      this.railH = rail.clientHeight || 200;
      const total = el.scrollHeight || 1;
      this.nodeTops = this.posts.map((p) => ((this.nodeEl(p.id)?.offsetTop || 0) / total) * 100);
      const max = el.scrollHeight - el.clientHeight;
      this.prog = max > 0 ? el.scrollTop / max : 0;
    },
    onScroll() {
      const el = this.$refs.scroller;
      if (!el) return;
      this.$store.ui.headerTitle = el.scrollTop > 60 && this.topic ? this.topic.title : null;
      const st = el.scrollTop + 60;
      let cur = null;
      for (const p of this.posts) {
        const node = this.nodeEl(p.id);
        if (node && node.offsetTop <= st) cur = p.id;
      }
      this.activePostId = cur || this.posts[0]?.id || null;
      const max = el.scrollHeight - el.clientHeight;
      this.prog = max > 0 ? el.scrollTop / max : 0;
      this.maybeSnap();
    },
    thumbFromY(clientY) {
      const rail = this.$refs.rail;
      const el = this.$refs.scroller;
      if (!rail || !el) return;
      const rect = rail.getBoundingClientRect();
      const max = el.scrollHeight - el.clientHeight;
      if (max <= 0) return;
      const th = Math.max(rect.height * (el.clientHeight / el.scrollHeight), 18);
      const t = (clientY - rect.top - th / 2) / Math.max(rect.height - th, 1);
      el.scrollTop = Math.min(Math.max(t, 0), 1) * max;
    },
    railDown(e) {
      this.dragging = true;
      this.$refs.rail?.setPointerCapture?.(e.pointerId);
      this.thumbFromY(e.clientY);
    },
    railMove(e) { if (this.dragging) this.thumbFromY(e.clientY); },
    railUp() { this.dragging = false; this.maybeSnap(); },
    railWheel(e) {
      const el = this.$refs.scroller;
      if (el) el.scrollTop = Math.max(0, el.scrollTop + (e.deltaY || 0));
    },
    maybeSnap() {
      const el = this.$refs.scroller;
      if (!el || this.dragging) return;
      const cur = el.scrollTop;
      let best = null;
      let bestD = SNAP;
      for (const p of this.posts) {
        const node = this.nodeEl(p.id);
        if (!node) continue;
        const target = Math.max(node.offsetTop - 24, 0);
        const d = Math.abs(cur - target);
        if (d < bestD) { bestD = d; best = { pid: p.id, pos: target }; }
      }
      if (best) {
        if (this.snap && Math.abs(cur - this.snap.pos) < RELEASE) return;
        this.snap = best;
        el.scrollTop = best.pos;
      } else if (this.snap) {
        this.snap = null;
      }
    },
    jumpTo(id) { this.nodeEl(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' }); },

    // ── 话题操作 ─────────────────────────────────────
    needLogin() {
      if (this.$store.session.logged) return false;
      this.$go('/login?from=' + encodeURIComponent(`/topics/${this.id}`));
      return true;
    },
    async submitReply() {
      if (this.needLogin()) return;
      if (!this.reply.trim()) return;
      this.submitting = true;
      try {
        const p = await api.createPost(this.id, { raw: this.reply.trim() });
        this.posts = [...this.posts, p];
        this.reply = '';
        this.$store.ui.ok('回复成功');
        this.$nextTick(() => this.calibrate());
      } catch (e) {
        this.$store.ui.fail(e || '回复失败');
      } finally { this.submitting = false; }
    },
    removePost(id) { this.posts = this.posts.filter((p) => p.id !== id); },
    openReport(post) { this.reportTarget = post; this.reportReason = ''; this.reportOpen = true; },
    async submitReport() {
      if (!this.reportTarget) return;
      if (!this.reportReason.trim()) { this.$store.ui.warn('请填写举报理由'); return; }
      try {
        await api.reportPost(this.reportTarget.id, this.reportReason.trim());
        this.$store.ui.ok('举报已提交，感谢反馈');
      } catch (e) { this.$store.ui.fail(e || '举报失败'); }
      this.reportOpen = false;
      this.reportTarget = null;
      this.reportReason = '';
    },
    async toggleBookmark() {
      if (this.needLogin()) return;
      try {
        const d = await api.bookmarkTopic(this.id);
        this.topic.bookmarked = d.bookmarked;
      } catch (e) { this.$store.ui.fail(e || '操作失败'); }
    },
    async toggleFollow() {
      if (this.needLogin()) return;
      try {
        const d = await api.followTopic(this.id);
        this.topic.followed = d.followed;
      } catch (e) { this.$store.ui.fail(e || '操作失败'); }
    },
    async setFlag(fn, msg) {
      try {
        await fn(this.id);
        this.$store.ui.ok(msg);
        this.topic = await api.getTopic(this.id);
      } catch (e) { this.$store.ui.fail(e || '操作失败'); }
    },
    togglePin() {
      if (!this.canEditTopic) return;
      this.setFlag(this.topic.pinned ? api.unpinTopic : api.pinTopic, this.topic.pinned ? '已取消置顶' : '已置顶');
    },
    toggleClose() {
      if (!this.canEditTopic) return;
      this.setFlag(this.topic.closed ? api.openTopic : api.closeTopic, this.topic.closed ? '已解锁' : '已锁定');
    },
    async toggleHide() {
      if (!this.canEditTopic) return;
      const next = !this.topic.hidden;
      if (next && !window.confirm('隐藏话题后前台将不再显示，确定隐藏？')) return;
      try {
        await api.hideTopic(this.id, next);
        this.topic.hidden = next;
        this.$store.ui.ok(next ? '已隐藏，列表不再显示' : '已恢复显示');
      } catch (e) { this.$store.ui.fail(e || '操作失败'); }
    },
    openEdit() {
      this.eTitle = this.topic?.title || '';
      this.eCat = this.topic?.category_id ? String(this.topic.category_id) : '';
      this.eTags = (this.topic?.tags || []).map((x) => x.name);
      this.editOpen = true;
    },
    setETags(v) { this.eTags = v; },
    async saveEdit() {
      try {
        this.topic = await api.updateTopic(this.id, {
          title: this.eTitle.trim(),
          category_id: this.eCat ? Number(this.eCat) : null,
          tags: this.eTags,
        });
        this.editOpen = false;
        this.$store.ui.ok('话题已更新');
      } catch (e) { this.$store.ui.fail(e || '更新失败'); }
    },
    toBack() {
      const back = this.$store.route.query.back;
      if (back) { this.$go('/?' + back); return; }
      this.$router.back();
    },

    badge: badgeStyle,
    initial,
    formatDate,
    formatDateTime,
    levelOf,
    levelStyle(user) {
      const c = safeColor(levelOf(user).color, '#8a9099');
      return `color:${c};background:${c}14;`;
    },
  };
}

/** 单个楼层：共享父作用域的 post 对象，删除/举报通过事件交回 topic 页面。 */
export function postRow(p) {
  return {
    editing: false,
    draft: p.raw || '',
    get isOp() { return p.post_number === 1; },
    get canEdit() {
      const s = this.$store.session;
      return !!s.user && (s.isStaff || s.user.username === p.user?.username);
    },
    startEdit() { this.draft = p.raw || ''; this.editing = true; },
    async save() {
      try {
        const d = await api.updatePost(p.id, { raw: this.draft });
        p.raw = d.raw;
        this.editing = false;
        this.$store.ui.ok('已更新');
      } catch (e) { this.$store.ui.fail(e || '保存失败'); }
    },
    remove() {
      if (!window.confirm('确定删除这条回复？')) return;
      this.$dispatch('post-delete', { id: p.id });
    },
    report() { this.$dispatch('post-report', { post: p }); },
    async vote(v) {
      if (!this.$store.session.logged) {
        this.$go('/login?from=' + encodeURIComponent(`/topics/${p.topic_id}`));
        return;
      }
      try {
        const d = await api.votePost(p.id, p.my_vote === v ? 0 : v);
        p.score = d.score;
        p.my_vote = d.my_vote;
        p.like_count = d.like_count;
      } catch (e) { this.$store.ui.fail(e || '投票失败'); }
    },
    initial,
    formatDateTime,
    levelOf,
    levelStyle(user) {
      const c = safeColor(levelOf(user).color, '#8a9099');
      return `color:${c};background:${c}14;`;
    },
  };
}
