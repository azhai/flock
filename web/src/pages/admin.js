import { api } from '../core/api.js';
import { paginationMixin, REPORT_STATUS_META, localeTime, safeColor } from '../core/utils.js';

export function admin() {
  return {
    ...paginationMixin,
    tab: 'overview',
    stats: null,
    cats: [],
    newCat: { name: '', slug: '', description: '' },
    reports: [],
    reportStatus: '',
    topics: [],
    topicHidden: '',
    topicPage: 1,
    topicTotal: 0,
    users: [],
    uq: '',
    userPage: 1,
    userTotal: 0,
    perPage: 20,

    // 两个分页 tab 共用 pagination 片段：按当前 tab 映射到各自的页码/总数。
    get page() { return this.tab === 'topics' ? this.topicPage : this.userPage; },
    set page(v) { if (this.tab === 'topics') this.topicPage = v; else this.userPage = v; },
    get total() { return this.tab === 'topics' ? this.topicTotal : this.userTotal; },

    init() {
      if (!this.$store.session.logged) { this.$go('/login?from=/admin'); return; }
      this.load();
    },
    setTab(name) { this.tab = name; this.load(); },
    load() {
      switch (this.tab) {
        case 'overview': this.loadStats(); break;
        case 'categories': this.loadCategories(); break;
        case 'reports': this.loadReports(); break;
        case 'topics': this.loadTopics(); break;
        case 'users': this.loadUsers(); break;
      }
    },

    statCards() {
      const s = this.stats || {};
      return [
        { label: '用户', value: s.users || 0, color: '#2563eb', to: 'users' },
        { label: '分类', value: s.cats || 0, color: '#2f855a', to: 'categories' },
        { label: '话题', value: s.topics || 0, color: '#d97706', to: 'topics' },
        { label: '待处理举报', value: s.reports || 0, color: '#dc2626', to: 'reports' },
      ];
    },
    async loadStats() {
      try {
        const [us, cs, ts, rs] = await Promise.all([
          api.listUsers(), api.listCategories(), api.listTopics({}), api.listReports('pending'),
        ]);
        this.stats = {
          users: us?.total || 0,
          cats: Array.isArray(cs) ? cs.length : 0,
          topics: ts?.total || 0,
          reports: rs?.total || 0,
        };
      } catch { /* 概览失败保持占位 */ }
    },

    loadCategories() {
      api.listCategories().then((d) => { this.cats = Array.isArray(d) ? d : []; }).catch(() => {});
    },
    async createCat() {
      if (!this.newCat.name.trim()) return;
      try {
        const c = await api.createCategory(this.newCat);
        this.cats = [...this.cats, c];
        this.newCat = { name: '', slug: '', description: '' };
        this.$store.ui.ok('创建成功');
      } catch (e) { this.$store.ui.fail(e || '创建失败'); }
    },
    async delCat(c) {
      if (!window.confirm(`删除分类「${c.name}」？`)) return;
      try {
        await api.deleteCategory(c.id);
        this.cats = this.cats.filter((x) => x.id !== c.id);
        this.$store.ui.ok('已删除');
      } catch (e) { this.$store.ui.fail(e || '删除失败'); }
    },

    loadReports() {
      api.listReports(this.reportStatus || undefined)
        .then((d) => { this.reports = d?.data || []; })
        .catch(() => {});
    },
    async actReport(r, action) {
      try {
        await (action === 'resolve' ? api.resolveReport(r.id) : api.dismissReport(r.id));
        this.$store.ui.ok(action === 'resolve' ? '已处理' : '已驳回');
        this.loadReports();
      } catch (e) { this.$store.ui.fail(e || '操作失败'); }
    },
    statusLabel(s) { return (REPORT_STATUS_META[s] || REPORT_STATUS_META.pending).label; },
    statusStyle(s) {
      const c = (REPORT_STATUS_META[s] || REPORT_STATUS_META.pending).color;
      return `color:${c};background:${c}14;`;
    },

    loadTopics() {
      api.listTopics({ sort: 'latest', hidden: this.topicHidden || undefined, page: this.topicPage })
        .then((d) => { this.topics = d?.data || []; this.topicTotal = d?.total || 0; })
        .catch(() => {});
    },
    async toggleHide(t) {
      const next = !t.hidden;
      if (next && !window.confirm(`隐藏话题「${t.title}」？隐藏后前台不再显示。`)) return;
      try {
        await api.hideTopic(t.id, next);
        this.$store.ui.ok(next ? '已隐藏' : '已恢复显示');
        this.loadTopics();
      } catch (e) { this.$store.ui.fail(e || '操作失败'); }
    },
    async togglePin(t) {
      try {
        await (t.pinned ? api.unpinTopic(t.id) : api.pinTopic(t.id));
        this.$store.ui.ok(t.pinned ? '已取消置顶' : '已置顶');
        this.loadTopics();
      } catch (e) { this.$store.ui.fail(e || '操作失败'); }
    },
    async toggleClose(t) {
      try {
        await (t.closed ? api.openTopic(t.id) : api.closeTopic(t.id));
        this.$store.ui.ok(t.closed ? '已解锁' : '已锁定');
        this.loadTopics();
      } catch (e) { this.$store.ui.fail(e || '操作失败'); }
    },
    async delTopic(t) {
      if (!window.confirm(`删除话题「${t.title}」？`)) return;
      try {
        await api.deleteTopic(t.id);
        this.$store.ui.ok('已删除');
        this.loadTopics();
      } catch (e) { this.$store.ui.fail(e || '删除失败'); }
    },

    loadUsers() {
      api.listUsers({ q: this.uq || undefined, page: this.userPage })
        .then((d) => { this.users = d?.data || []; this.userTotal = d?.total || 0; })
        .catch(() => {});
    },
    async saveUser(u, patch) {
      try {
        await api.adminUpdateUser(u.username, patch);
        this.$store.ui.ok('已保存');
        this.loadUsers();
      } catch (e) { this.$store.ui.fail(e || '保存失败'); }
    },

    dot(color) { return 'background:' + safeColor(color, '#718096') + ';'; },
    localeTime,
  };
}
