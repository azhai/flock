import { api } from './api.js';

function readUser() {
  try { return JSON.parse(localStorage.getItem('user') || 'null'); }
  catch { return null; }
}

function messageOf(err) {
  if (err == null) return '操作失败';
  if (typeof err === 'string') return err;
  return String(err.message || err.error || err);
}

export function registerStores(Alpine) {
  Alpine.store('session', {
    user: readUser(),
    token: localStorage.getItem('token') || '',
    get logged() { return !!this.user && !!this.token; },
    get isStaff() { return this.user?.role === 'admin' || this.user?.role === 'moderator'; },
    setSession(user, token) {
      this.user = user;
      this.token = token;
      localStorage.setItem('user', JSON.stringify(user));
      localStorage.setItem('token', token);
      Alpine.store('notifications').start();
    },
    logout() {
      this.user = null;
      this.token = '';
      localStorage.removeItem('user');
      localStorage.removeItem('token');
      Alpine.store('notifications').stop();
    },
  });

  Alpine.store('ui', {
    toasts: [],
    headerTitle: null,
    toast(message, status = 'info') {
      const id = Date.now() + Math.random();
      this.toasts.push({ id, message: messageOf(message), status });
      setTimeout(() => {
        this.toasts = this.toasts.filter((t) => t.id !== id);
      }, 3200);
    },
    ok(message) { this.toast(message, 'success'); },
    fail(message) { this.toast(message, 'error'); },
    warn(message) { this.toast(message, 'warning'); },
  });

  Alpine.store('route', {
    name: 'home', params: {}, query: {}, search: '', path: '/', version: 0,
  });

  Alpine.store('notifications', {
    unread: 0,
    list: [],
    _timer: null,
    _abort: null,
    _running: false,

    start() {
      if (this._running || !Alpine.store('session').logged) return;
      this._running = true;
      this.refresh();
      this._timer = setInterval(() => this.tick(), 30000);
      this.stream();
    },
    stop() {
      this._running = false;
      if (this._timer) clearInterval(this._timer);
      this._timer = null;
      if (this._abort) this._abort.abort();
      this._abort = null;
      this.unread = 0;
      this.list = [];
    },
    async tick() {
      try { const d = await api.unreadCount(); this.unread = d.unread || 0; }
      catch { /* 忽略：轮询失败不影响页面 */ }
    },
    async refresh() {
      try { const list = await api.listNotifications(); this.list = Array.isArray(list) ? list : []; }
      catch { /* 忽略 */ }
      await this.tick();
    },
    // SSE：EventSource 无法携带 Authorization，沿用 fetch 流；断开时保留 30s 轮询兜底。
    async stream() {
      const ctrl = new AbortController();
      this._abort = ctrl;
      try {
        const res = await fetch('/api/v1/notifications/stream', {
          headers: { Authorization: `Bearer ${Alpine.store('session').token}` },
          signal: ctrl.signal,
        });
        if (!res.ok || !res.body) return;
        const reader = res.body.getReader();
        const decoder = new TextDecoder();
        let buf = '';
        for (;;) {
          const { value, done } = await reader.read();
          if (done) break;
          buf += decoder.decode(value, { stream: true });
          const frames = buf.split('\n\n');
          buf = frames.pop() || '';
          for (const frame of frames) {
            const m = /data:\s*(.*)/.exec(frame);
            if (!m) continue;
            try {
              const d = JSON.parse(m[1]);
              if (typeof d.unread === 'number') this.unread = d.unread;
              if (d.unread !== undefined) {
                const list = await api.listNotifications();
                this.list = Array.isArray(list) ? list : [];
              }
            } catch { /* 单个帧解析失败忽略 */ }
          }
        }
      } catch {
        // 断开/不支持时轮询仍在跑，无需额外处理
      }
    },
    async markRead(id) {
      try {
        await api.markNotificationRead(id);
        const n = this.list.find((x) => x.id === id);
        if (n && !n.read) { n.read = true; this.unread = Math.max(0, this.unread - 1); }
      } catch { /* 忽略 */ }
    },
    async markAll() {
      try {
        await api.markAllRead();
        this.list.forEach((n) => { n.read = true; });
        this.unread = 0;
      } catch { /* 忽略 */ }
    },
  });
}
