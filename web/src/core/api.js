const BASE = '/api/v1';

async function req(method, path, body) {
  const token = localStorage.getItem('token');
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers.Authorization = `Bearer ${token}`;
  const opts = { method, headers, credentials: 'omit' };
  if (body !== undefined) opts.body = JSON.stringify(body);

  const res = await fetch(BASE + path, opts);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw data.error || res.statusText;
  return data;
}

function qs(params = {}) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') q.set(k, v);
  }
  const s = q.toString();
  return s ? '?' + s : '';
}

export const api = {
  // auth
  register: (r) => req('POST', '/auth/register', r),
  login: (r) => req('POST', '/auth/login', r),
  me: () => req('GET', '/auth/me'),

  // categories / tags
  listCategories: () => req('GET', '/categories'),
  listTags: () => req('GET', '/tags'),
  createCategory: (r) => req('POST', '/categories', r),
  updateCategory: (id, r) => req('PUT', `/categories/${id}`, r),
  deleteCategory: (id) => req('DELETE', `/categories/${id}`),

  // topics
  listTopics: (params = {}) => req('GET', '/topics' + qs(params)),
  getTopic: (id) => req('GET', `/topics/${id}`),
  createTopic: (r) => req('POST', '/topics', r),
  updateTopic: (id, r) => req('PUT', `/topics/${id}`, r),
  deleteTopic: (id) => req('DELETE', `/topics/${id}`),
  pinTopic: (id) => req('POST', `/topics/${id}/pin`),
  unpinTopic: (id) => req('POST', `/topics/${id}/unpin`),
  closeTopic: (id) => req('POST', `/topics/${id}/close`),
  openTopic: (id) => req('POST', `/topics/${id}/open`),
  hideTopic: (id, hidden) => req('POST', `/topics/${id}/hide`, { hidden }),

  // posts
  listPosts: (topicId, page) => req('GET', `/topics/${topicId}/posts` + qs({ page })),
  createPost: (topicId, r) => req('POST', `/topics/${topicId}/posts`, r),
  updatePost: (id, r) => req('PUT', `/posts/${id}`, r),
  deletePost: (id) => req('DELETE', `/posts/${id}`),
  toggleLike: (id) => req('POST', `/posts/${id}/like`),
  votePost: (id, vote) => req('POST', `/posts/${id}/vote`, { vote }),
  reportPost: (id, reason) => req('POST', `/posts/${id}/report`, { reason }),

  // notifications
  listNotifications: () => req('GET', '/notifications'),
  unreadCount: () => req('GET', '/notifications/unread'),
  markNotificationRead: (id) => req('POST', `/notifications/${id}/read`),
  markAllRead: () => req('POST', '/notifications/read-all'),

  // bookmarks / follow
  bookmarkTopic: (id) => req('POST', `/topics/${id}/bookmark`),
  listBookmarks: () => req('GET', '/bookmarks'),
  followTopic: (id) => req('POST', `/topics/${id}/follow`),

  // users
  getUser: (username) => req('GET', `/users/${encodeURIComponent(username)}`),
  updateProfile: (r) => req('PUT', '/user/me', r),
  getUserTopics: (username, page) =>
    req('GET', `/users/${encodeURIComponent(username)}/topics` + qs({ page })),
  getUserPosts: (username, page) =>
    req('GET', `/users/${encodeURIComponent(username)}/posts` + qs({ page })),

  // admin
  listUsers: (params = {}) => req('GET', '/admin/users' + qs(params)),
  adminUpdateUser: (username, r) => req('PUT', `/admin/users/${encodeURIComponent(username)}`, r),
  listReports: (status) => req('GET', '/reports' + qs({ status })),
  resolveReport: (id) => req('POST', `/reports/${id}/resolve`),
  dismissReport: (id) => req('POST', `/reports/${id}/dismiss`),
};
