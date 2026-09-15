import m from 'mithril';
import { api } from './api';

let started = false;
const state = { unread: 0, list: [] as any[] };

function tick() {
  api.unreadCount().then((d: any) => {
    state.unread = d.unread || 0;
    m.redraw();
  }).catch(() => {});
}

function loadList() {
  api.listNotifications().then((list: any) => {
    state.list = list || [];
    m.redraw();
  }).catch(() => {});
}

function startSSE() {
  const token = localStorage.getItem('token');
  fetch('/api/v1/notifications/stream', {
    headers: { Authorization: `Bearer ${token}` },
  }).then((res) => {
    if (!res.ok || !res.body) {
      setInterval(tick, 30000);
      return;
    }
    const reader = res.body.getReader();
    const dec = new TextDecoder();
    let buf = '';
    (async () => {
      for (;;) {
        const { value, done } = await reader.read();
        if (done) break;
        buf += dec.decode(value, { stream: true });
        const frames = buf.split('\n\n');
        buf = frames.pop() || '';
        for (const f of frames) {
          const mm = f.match(/data:\s*(.*)/);
          if (!mm) continue;
          try {
            const d = JSON.parse(mm[1]);
            if (typeof d.unread === 'number') state.unread = d.unread;
            if (d.unread !== undefined) loadList();
            m.redraw();
          } catch {}
        }
      }
    })();
  }).catch(() => {
    setInterval(tick, 30000);
  });
}

export const notify = {
  get state() {
    return state;
  },
  start() {
    if (started) return;
    started = true;
    loadList();
    setInterval(tick, 30000);
    startSSE();
  },
  stop() {
    started = false;
    state.unread = 0;
    state.list = [];
    m.redraw();
  },
  markRead(id: number) {
    api.markNotificationRead(id).then(() => {
      state.list = state.list.map((n) => (n.id === id ? { ...n, read: true } : n));
      state.unread = Math.max(0, state.unread - 1);
      m.redraw();
    }).catch(() => {});
  },
  markAll() {
    api.markAllRead().then(() => {
      state.list = state.list.map((n) => ({ ...n, read: true }));
      state.unread = 0;
      m.redraw();
    }).catch(() => {});
  },
};
