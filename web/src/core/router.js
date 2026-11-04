// 路由 seam：正则路由表 + History API + 站内 <a> 点击委托。
// 模板里直接用普通链接（可中键打开、可复制链接），无需 @click 包装。
const ROUTES = [
  { name: 'home', re: /^\/$/ },
  { name: 'topic', re: /^\/topics\/(?<id>\d+)\/?$/ },
  { name: 'new', re: /^\/new\/?$/ },
  { name: 'login', re: /^\/login\/?$/ },
  { name: 'register', re: /^\/register\/?$/ },
  { name: 'profile', re: /^\/u\/(?<username>[^/]+)\/?$/ },
  { name: 'bookmarks', re: /^\/bookmarks\/?$/ },
  { name: 'admin', re: /^\/admin\/?$/ },
];

export function createRouter(Alpine) {
  const store = () => Alpine.store('route');

  function parse(url) {
    for (const r of ROUTES) {
      const m = r.re.exec(url.pathname);
      if (!m) continue;
      const params = {};
      for (const [k, v] of Object.entries(m.groups || {})) {
        try { params[k] = decodeURIComponent(v); } catch { params[k] = v; }
      }
      return {
        name: r.name,
        params,
        query: Object.fromEntries(url.searchParams.entries()),
        search: url.search,
        path: url.pathname + url.search,
      };
    }
    return null;
  }

  function apply() {
    let url = new URL(location.href);
    let matched = parse(url);
    if (!matched) {
      history.replaceState({}, '', '/');
      url = new URL(location.href);
      matched = parse(url);
    }
    const s = store();
    const prevName = s.name;
    Object.assign(s, { ...matched, version: s.version + 1 });
    if (s.name !== 'topic') Alpine.store('ui').headerTitle = null;
    if (prevName !== s.name) window.scrollTo(0, 0);
  }

  function navigate(to, { replace = false } = {}) {
    const target = new URL(to, location.href);
    // 只接受同源路径，避免 URL 参数被构造成开放重定向。
    if (target.origin !== location.origin) return;
    const path = target.pathname + target.search + target.hash;
    history[replace ? 'replaceState' : 'pushState']({}, '', path);
    apply();
  }

  function onClick(e) {
    if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
    const a = e.target?.closest?.('a[href]');
    if (!a) return;
    if (a.target && a.target !== '_self') return;
    if (a.hasAttribute('download')) return;
    const raw = a.getAttribute('href') || '';
    if (raw.startsWith('#') || raw.startsWith('mailto:') || raw.startsWith('tel:')) return;
    const url = new URL(a.href, location.href);
    if (url.origin !== location.origin || url.pathname.startsWith('/api/')) return;
    e.preventDefault();
    if (url.pathname === location.pathname && url.search === location.search) return;
    navigate(url.pathname + url.search + url.hash);
  }

  return {
    go: navigate,
    back() { if (window.history.length > 1) history.back(); else navigate('/'); },
    start() {
      document.addEventListener('click', onClick);
      window.addEventListener('popstate', apply);
      apply();
    },
  };
}
