import { api } from '../core/api.js';
import { SORTS, timeAgo, initial, badgeStyle, paginationMixin } from '../core/utils.js';

export function home() {
  return {
    ...paginationMixin,
    cats: [],
    topics: [],
    total: 0,
    page: 1,
    perPage: 20,
    loading: true,
    sorts: SORTS,

    get q() { return (this.$store.route.query.q || '').trim(); },
    get catSlug() { return this.$store.route.query.cat || ''; },
    get tag() { return this.$store.route.query.tag || ''; },
    get sort() { return this.$store.route.query.sort || 'latest'; },
    get heading() {
      if (this.q) return `搜索“${this.q}”`;
      const c = this.cats.find((x) => x.slug === this.catSlug);
      if (c) return c.name;
      return this.tag ? `#${this.tag}` : '最新话题';
    },

    async init() {
      this.$watch('$store.route.version', () => {
        this.page = Number(this.$store.route.query.page) || 1;
        this.load();
      });
      try {
        const d = await api.listCategories();
        this.cats = Array.isArray(d) ? d : [];
      } catch {
        this.$store.ui.fail('加载分类失败');
      }
      this.page = Number(this.$store.route.query.page) || 1;
      this.load();
    },

    load() {
      const catId = this.cats.find((c) => c.slug === this.catSlug)?.id;
      this.loading = true;
      api.listTopics({
        category_id: catId || undefined,
        tag: this.tag || undefined,
        q: this.q || undefined,
        sort: this.sort === 'latest' ? undefined : this.sort,
        page: this.page,
      })
        .then((d) => { this.topics = d?.data || []; this.total = d?.total || 0; })
        .catch((e) => this.$store.ui.fail(e || '加载话题失败'))
        .finally(() => { this.loading = false; });
    },

    setSort(key) {
      const q = new URLSearchParams();
      if (this.q) q.set('q', this.q);
      if (this.catSlug) q.set('cat', this.catSlug);
      if (this.tag) q.set('tag', this.tag);
      if (key !== 'latest') q.set('sort', key);
      this.$go('/' + (q.toString() ? '?' + q : ''));
    },

    topicHref(t) {
      const back = new URLSearchParams();
      if (this.catSlug) back.set('cat', this.catSlug);
      if (this.tag) back.set('tag', this.tag);
      if (this.sort !== 'latest') back.set('sort', this.sort);
      if (this.page > 1) back.set('page', this.page);
      const qs = back.toString();
      return `/topics/${t.id}` + (qs ? '?back=' + encodeURIComponent(qs) : '');
    },

    participants(t) {
      return t.participants?.length ? t.participants : (t.user ? [t.user] : []);
    },
    timeAgo,
    initial,
    badge: badgeStyle,
  };
}
