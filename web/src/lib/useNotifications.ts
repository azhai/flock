import { useCallback, useEffect, useRef, useState } from 'react';
import { api } from './api';

// 通知订阅：优先用 fetch 流式读 SSE（可带 Authorization 头），失败/断开回退轮询
export function useNotifications(active: boolean) {
  const [unread, setUnread] = useState(0);
  const [notifs, setNotifs] = useState<any[]>([]);
  const pollTimer = useRef<any>(null);
  const cancelled = useRef(false);

  const tick = useCallback(async () => {
    try {
      const d = await api.unreadCount();
      if (!cancelled.current) setUnread(d.unread || 0);
    } catch { /* ignore */ }
  }, []);

  const loadList = useCallback(async () => {
    try {
      const list = await api.listNotifications();
      if (!cancelled.current) setNotifs(list || []);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => {
    if (!active) return;
    cancelled.current = false;
    (async () => {
      await loadList();
      pollTimer.current = setInterval(tick, 30000);
      const token = localStorage.getItem('token');
      // SSE 流（fetch 版，带认证头）
      try {
        const res = await fetch('/api/v1/notifications/stream', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!res.ok || !res.body || cancelled.current) return;
        const reader = res.body.getReader();
        const dec = new TextDecoder();
        let buf = '';
        for (;;) {
          if (cancelled.current) break;
          const { value, done } = await reader.read();
          if (done) break;
          buf += dec.decode(value, { stream: true });
          const frames = buf.split('\n\n');
          buf = frames.pop() || '';
          for (const frame of frames) {
            const m = frame.match(/data:\s*(.*)/);
            if (!m) continue;
            try {
              const d = JSON.parse(m[1]);
              if (typeof d.unread === 'number' && !cancelled.current) setUnread(d.unread);
              if (d.unread !== undefined) loadList(); // 有新通知就刷新列表
            } catch { /* ignore */ }
          }
        }
      } catch {
        clearInterval(pollTimer.current);
        pollTimer.current = setInterval(tick, 30000); // SSE 失败则纯轮询
      }
    })();
    return () => {
      cancelled.current = true;
      if (pollTimer.current) clearInterval(pollTimer.current);
    };
  }, [active, loadList, tick]);

  const markRead = useCallback(async (id: number) => {
    try {
      await api.markNotificationRead(id);
      setNotifs((ns) => ns.map((n) => (n.id === id ? { ...n, read: true } : n)));
      setUnread((u) => Math.max(0, u - 1));
    } catch { /* ignore */ }
  }, []);

  const markAll = useCallback(async () => {
    try {
      await api.markAllRead();
      setNotifs((ns) => ns.map((n) => ({ ...n, read: true })));
      setUnread(0);
    } catch { /* ignore */ }
  }, []);

  return { unread, notifs, markRead, markAll, refresh: loadList };
}