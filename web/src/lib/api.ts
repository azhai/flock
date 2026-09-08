const BASE = '/api/v1';

async function req(method, path, body) {
  const token = localStorage.getItem('token');
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
    credentials: 'omit',
  };
  if (token) opts.headers['Authorization'] = `Bearer ${token}`;
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(BASE + path, opts);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw data.error || res.statusText;
  return data;
}

export const api = {
  // auth
  register: (r) => req('POST', '/auth/register', r),
  login: (r) => req('POST', '/auth/login', r),
  me: () => req('GET', '/auth/me'),

  // categories
  listCategories: () => req('GET', '/categories'),
  listTags: () => req('GET', '/tags'),
  createCategory: (r) => req('POST', '/categories', r),
  updateCategory: (id, r) => req('PUT', `/categories/${id}`, r),
  deleteCategory: (id) => req('DELETE', `/categories/${id}`),

  // topics
  listTopics: (params) => {
    const q = new URLSearchParams();
    if (params?.category_id) q.set('category_id', params.category_id);
    if (params?.q) q.set('q', params.q);
    if (params?.tag) q.set('tag', params.tag);
    if (params?.sort) q.set('sort', params.sort);
    if (params?.page) q.set('page', params.page);
    return req('GET', `/topics?${q.toString()}`);
  },
  getTopic: (id) => req('GET', `/topics/${id}`),
  createTopic: (r) => req('POST', '/topics', r),
  updateTopic: (id, r) => req('PUT', `/topics/${id}`, r),
  deleteTopic: (id) => req('DELETE', `/topics/${id}`),
  pinTopic: (id) => req('POST', `/topics/${id}/pin`),
  unpinTopic: (id) => req('POST', `/topics/${id}/unpin`),
  closeTopic: (id) => req('POST', `/topics/${id}/close`),
  openTopic: (id) => req('POST', `/topics/${id}/open`),

  // posts
  listPosts: (topicId, page) =>
    req('GET', `/topics/${topicId}/posts${page ? '?page=' + page : ''}`),
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
  bookmarkTopic: (topicId) => req('POST', `/topics/${topicId}/bookmark`),
  listBookmarks: () => req('GET', '/bookmarks'),
  followTopic: (topicId) => req('POST', `/topics/${topicId}/follow`),

  // users
  getUser: (username) => req('GET', `/users/${username}`),
  updateProfile: (r) => req('PUT', '/user/me', r),
  getUserTopics: (username, page) =>
    req('GET', `/users/${username}/topics${page ? '?page=' + page : ''}`),
  getUserPosts: (username, page) =>
    req('GET', `/users/${username}/posts${page ? '?page=' + page : ''}`),

  // admin
  listUsers: (params = {}) => {
    const q = new URLSearchParams();
    if (params.q) q.set('q', params.q);
    if (params.page) q.set('page', params.page);
    return req('GET', `/admin/users?${q.toString()}`);
  },
  adminUpdateUser: (username, r) => req('PUT', `/admin/users/${username}`, r),
  listReports: (status) => req('GET', `/reports${status ? '?status=' + status : ''}`),
  resolveReport: (id) => req('POST', `/reports/${id}/resolve`),
  dismissReport: (id) => req('POST', `/reports/${id}/dismiss`),
};