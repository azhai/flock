import { api } from '../core/api.js';
import { editorMixin, EMOJIS } from '../core/utils.js';

export function newTopic() {
  return {
    ...editorMixin,
    emojis: EMOJIS,
    title: '',
    categoryId: '',
    tags: [],
    raw: '',
    cats: [],
    submitting: false,

    init() {
      api.listCategories()
        .then((d) => { this.cats = Array.isArray(d) ? d : []; })
        .catch(() => {});
    },
    setTags(v) { this.tags = v; },

    async submit() {
      if (!this.title.trim()) { this.$store.ui.warn('请填写标题'); return; }
      if (!this.raw.trim()) { this.$store.ui.warn('请填写内容'); return; }
      if (!this.$store.session.logged) { this.$go('/login?from=/new'); return; }
      this.submitting = true;
      try {
        const t = await api.createTopic({
          title: this.title.trim(),
          raw: this.raw.trim(),
          category_id: this.categoryId ? Number(this.categoryId) : null,
          tags: this.tags,
        });
        this.$store.ui.ok('发布成功');
        this.$go(`/topics/${t.id || t.topic?.id}`);
      } catch (e) {
        this.$store.ui.fail(e || '发布失败');
      } finally {
        this.submitting = false;
      }
    },
  };
}
