import { api } from './core/api.js';
import { initial, timeAgo, NOTIF_META, safeColor } from './core/utils.js';

export function registerComponents(Alpine) {
  Alpine.data('siteHeader', () => ({
    q: '',
    notifOpen: false,
    userOpen: false,
    get unread() { return this.$store.notifications.unread; },
    get notifs() { return this.$store.notifications.list; },

    submitSearch() {
      const kw = this.q.trim();
      this.$go(kw ? '/?q=' + encodeURIComponent(kw) : '/');
    },
    meta(type) { return NOTIF_META[type] || NOTIF_META.default; },
    timeAgo,
    initial,
    openNotif(n) {
      if (!n.read) this.$store.notifications.markRead(n.id);
      this.notifOpen = false;
      if (n.topic_id) this.$go(`/topics/${n.topic_id}`);
    },
    markAll() { this.$store.notifications.markAll(); },
    logout() {
      this.$store.session.logout();
      this.userOpen = false;
      this.$go('/');
    },
  }));

  Alpine.data('sidebar', () => ({
    cats: [],
    tags: [],
    init() {
      api.listCategories().then((d) => { this.cats = Array.isArray(d) ? d : []; }).catch(() => {});
      api.listTags().then((d) => { this.tags = Array.isArray(d) ? d : []; }).catch(() => {});
    },
    get activeCat() { return this.$store.route.query.cat || ''; },
    get activeTag() { return this.$store.route.query.tag || ''; },
    selectCat(slug) {
      const q = new URLSearchParams();
      if (slug) q.set('cat', slug);
      if (this.activeTag) q.set('tag', this.activeTag);
      this.$go('/' + (q.toString() ? '?' + q : ''));
    },
    selectTag(slug) {
      const q = new URLSearchParams();
      if (this.activeCat) q.set('cat', this.activeCat);
      if (slug) q.set('tag', slug);
      this.$go('/' + (q.toString() ? '?' + q : ''));
    },
    dot(color) { return 'background:' + safeColor(color, '#718096') + ';'; },
  }));

  // 用法：x-data="tagInput(() => tags, (v) => tags = v)"
  // 通过 getter/setter 与宿主页面双向同步，避免 Alpine 跨作用域绑定。
  Alpine.data('tagInput', (get, set) => ({
    draft: '',
    suggestions: [],
    init() {
      api.listTags()
        .then((d) => { this.suggestions = Array.isArray(d) ? d : []; })
        .catch(() => {});
    },
    get value() { return get() || []; },
    get matches() {
      const d = this.draft.trim().toLowerCase();
      if (!d) return [];
      return this.suggestions
        .filter((s) => String(s.name).includes(d) && !this.value.some((t) => t.toLowerCase() === s.name))
        .slice(0, 5);
    },
    commit(name) {
      const n = String(name || '').trim().replace(/^#/, '');
      if (!n) return;
      const v = this.value;
      if (!v.some((t) => t.toLowerCase() === n.toLowerCase()) && v.length < 8) set([...v, n]);
      this.draft = '';
    },
    onKey(e) {
      if (e.key === 'Enter' || e.key === ',' || e.key === ' ') {
        e.preventDefault();
        this.commit(this.draft);
      } else if (e.key === 'Backspace' && !this.draft && this.value.length) {
        set(this.value.slice(0, -1));
      }
    },
    remove(t) { set(this.value.filter((x) => x !== t)); },
  }));
}
