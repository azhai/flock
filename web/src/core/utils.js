export function timeAgo(dateStr) {
  if (!dateStr) return '';
  const ms = Date.now() - new Date(dateStr).getTime();
  if (!Number.isFinite(ms) || ms < 0) return '刚刚';
  if (ms < 60000) return '刚刚';
  if (ms < 3600000) return `${Math.floor(ms / 60000)}分钟前`;
  if (ms < 86400000) return `${Math.floor(ms / 3600000)}小时前`;
  if (ms < 604800000) return `${Math.floor(ms / 86400000)}天前`;
  return new Date(dateStr).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
}

export function formatDate(dateStr) {
  if (!dateStr) return '';
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' });
}

export function formatDateTime(dateStr) {
  if (!dateStr) return '';
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  });
}

export function localeTime(dateStr) {
  if (!dateStr) return '';
  return new Date(dateStr).toLocaleString('zh-CN');
}

export function initial(name = '') {
  return String(name || '?').trim().charAt(0).toUpperCase() || '?';
}

// 后端返回的颜色值会拼进 inline style：只接受 #rgb / #rrggbb / #rrggbbaa。
export function safeColor(color, fallback = '#0d8a6f') {
  return /^#(?:[0-9a-f]{3}|[0-9a-f]{6}|[0-9a-f]{8})$/i.test(String(color || ''))
    ? String(color).toLowerCase()
    : fallback;
}

export function badgeStyle(color, alpha = '1a') {
  const c = safeColor(color);
  const bg = c.length === 7 ? c + alpha : c;
  return `color:${c};background:${bg};`;
}

export const LEVEL_META = {
  0: { name: '新手', color: '#8a9099' },
  1: { name: '初级', color: '#2b6cb0' },
  2: { name: '中级', color: '#2f855a' },
  3: { name: '资深', color: '#d97706' },
};

export function levelOf(user) {
  return LEVEL_META[user?.level] || { name: user?.level_name || '新手', color: '#8a9099' };
}

export const NOTIF_META = {
  reply: { label: '回复了你的话题', icon: 'message', color: '#0d8a6f' },
  mention: { label: '提到了你', icon: 'at', color: '#d97706' },
  like: { label: '赞了你的回复', icon: 'thumb', color: '#dc2626' },
  follow: { label: '关注了你的话题', icon: 'user', color: '#2563eb' },
  report: { label: '提交了新举报', icon: 'flag', color: '#9333ea' },
  default: { label: '新动态', icon: 'bell', color: '#5f6571' },
};

export const REPORT_STATUS_META = {
  pending: { label: '待处理', color: '#d97706' },
  resolved: { label: '已处理', color: '#2f855a' },
  dismissed: { label: '已驳回', color: '#8a9099' },
};

export const SORTS = [
  { key: 'latest', label: '最新' },
  { key: 'hot', label: '活跃' },
  { key: 'top', label: '热门' },
  { key: 'likes', label: '获赞' },
];

export const EMOJIS = [
  '😀', '😄', '😂', '🤣', '😊', '😍', '🥰', '😎', '🤔', '😅',
  '🙌', '👏', '👍', '👎', '🙏', '💪', '🤝', '👋', '🎉', '🔥',
  '💯', '✅', '❌', '❤️', '💚', '💡', '❗', '❓', '🚀', '📌',
  '🏆', '📢', '👀', '💬', '🗃️', '⚡', '☁️', '🐛', '🧩', '📝',
];

/** 页号窗口：number 或 {ellipsis:true}，key 唯一。 */
export function buildPages(page, totalPages, siblings = 1) {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, i) => ({ key: String(i + 1), num: i + 1 }));
  }
  const start = Math.max(2, page - siblings);
  const end = Math.min(totalPages - 1, page + siblings);
  const out = [{ key: '1', num: 1 }];
  if (start > 2) out.push({ key: 'e1', ellipsis: true });
  for (let i = start; i <= end; i++) out.push({ key: String(i), num: i });
  if (end < totalPages - 1) out.push({ key: 'e2', ellipsis: true });
  out.push({ key: String(totalPages), num: totalPages });
  return out;
}

/** 混入需要分页的页面 data：提供 pageWindow / goPage / submitJump。 */
// 注意：mixin 会被对象展开（spread），getter 会退化成静态值，因此一律用方法。
export const paginationMixin = {
  jump: '',
  totalPages() { return Math.max(1, Math.ceil((this.total || 0) / this.perPage)); },
  pageWindow() { return buildPages(this.page, this.totalPages()); },
  goPage(p) {
    const n = Number(p);
    if (!Number.isFinite(n) || n < 1 || n > this.totalPages() || n === this.page) return;
    this.page = n;
    this.load();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  },
  submitJump() {
    const n = parseInt(this.jump, 10);
    if (!Number.isNaN(n)) this.goPage(n);
    this.jump = '';
  },
};

/** 混入带 Markdown 编辑器的页面 data（每页最多一个工具栏编辑器）。 */
export const editorMixin = {
  showEmoji: false,
  preview: false,
  toolbar(kind) {
    const ta = this.$refs.mdTextarea;
    if (!ta) return;
    const pairs = {
      bold: ['**', '**'], italic: ['*', '*'], code: ['`', '`'],
      link: ['[', '](https://)'], quote: ['> ', ''], list: ['- ', ''],
      heading: ['## ', ''], codeblock: ['```\n', '\n```'],
    };
    const [before, after] = pairs[kind] || ['', ''];
    const start = ta.selectionStart ?? 0;
    const end = ta.selectionEnd ?? 0;
    const selected = ta.value.slice(start, end);
    ta.focus();
    ta.setRangeText(before + selected + after, start, end, 'end');
    // 无选中文本时把光标放进标记中间，直接接着输入。
    if (!selected) ta.setSelectionRange(start + before.length, start + before.length);
    ta.dispatchEvent(new Event('input', { bubbles: true }));
  },
  addEmoji(e) {
    const ta = this.$refs.mdTextarea;
    if (!ta) return;
    const start = ta.selectionStart ?? ta.value.length;
    ta.setRangeText(e, start, start, 'end');
    ta.dispatchEvent(new Event('input', { bubbles: true }));
    this.showEmoji = false;
    ta.focus();
  },
};
