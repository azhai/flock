import { api } from '../core/api.js';
import { badgeStyle, initial, formatDate, levelOf } from '../core/utils.js';

export function bookmarks() {
  return {
    items: null,
    init() {
      if (!this.$store.session.logged) {
        this.$go('/login?from=/bookmarks');
        return;
      }
      api.listBookmarks()
        .then((d) => { this.items = Array.isArray(d) ? d : []; })
        .catch((e) => { this.items = []; this.$store.ui.fail(e || '加载收藏失败'); });
    },
    badge: badgeStyle,
  };
}

export function profile() {
  return {
    user: null,
    topics: [],
    init() {
      this.$watch('$store.route.version', () => this.load());
      this.load();
    },
    async load() {
      const username = this.$store.route.params.username || '';
      if (!username) return;
      try { this.user = await api.getUser(username); } catch { this.user = null; }
      try {
        const d = await api.getUserTopics(username);
        this.topics = d?.data || [];
      } catch { this.topics = []; }
    },
    initial,
    formatDate,
    levelOf,
  };
}
