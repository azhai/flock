const BASE = '/api/v1';

async function req(method: string, path: string, body?: any) {
  const token = localStorage.getItem('token');
  const opts: any = {
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
  register: (r: any) => req('POST', '/auth/register', r),
  login: (r: any) => req('POST', '/auth/login', r),
  me: () => req('GET', '/auth/me'),

  // categories
  listCategories: () => req('GET', '/categories'),
  listTags: () => req('GET', '/tags'),
  createCategory: (r: any) => req('POST', '/categories', r),
  updateCategory: (id: any, r: any) => req('PUT', `/categories/${id}`, r),
  deleteCategory: (id: any) => req('DELETE', `/categories/${id}`),

  // topics
  listTopics: (params: any = {}) => {
    const q = new URLSearchParams();
    if (params.category_id) q.set('category_id', params.category_id);
    if (params.q) q.set('q', params.q);
    if (params.tag) q.set('tag', params.tag);
    if (params.sort) q.set('sort', params.sort);
    if (params.page) q.set('page', params.page);
    if (params.hidden) q.set('hidden', params.hidden);
    return req('GET', `/topics?${q.toString()}`);
  },
  getTopic: (id: any) => req('GET', `/topics/${id}`),
  createTopic: (r: any) => req('POST', '/topics', r),
  updateTopic: (id: any, r: any) => req('PUT', `/topics/${id}`, r),
  deleteTopic: (id: any) => req('DELETE', `/topics/${id}`),
  pinTopic: (id: any) => req('POST', `/topics/${id}/pin`),
  unpinTopic: (id: any) => req('POST', `/topics/${id}/unpin`),
  closeTopic: (id: any) => req('POST', `/topics/${id}/close`),
  openTopic: (id: any) => req('POST', `/topics/${id}/open`),
  hideTopic: (id: any, hidden: any) => req('POST', `/topics/${id}/hide`, { hidden }),

  // posts
  listPosts: (topicId: any, page?: any) =>
    req('GET', `/topics/${topicId}/posts${page ? '?page=' + page : ''}`),
  createPost: (topicId: any, r: any) => req('POST', `/topics/${topicId}/posts`, r),
  updatePost: (id: any, r: any) => req('PUT', `/posts/${id}`, r),
  deletePost: (id: any) => req('DELETE', `/posts/${id}`),
  toggleLike: (id: any) => req('POST', `/posts/${id}/like`),
  votePost: (id: any, vote: any) => req('POST', `/posts/${id}/vote`, { vote }),
  reportPost: (id: any, reason: any) => req('POST', `/posts/${id}/report`, { reason }),

  // notifications
  listNotifications: () => req('GET', '/notifications'),
  unreadCount: () => req('GET', '/notifications/unread'),
  markNotificationRead: (id: any) => req('POST', `/notifications/${id}/read`),
  markAllRead: () => req('POST', '/notifications/read-all'),

  // bookmarks / follow
  bookmarkTopic: (topicId: any) => req('POST', `/topics/${topicId}/bookmark`),
  listBookmarks: () => req('GET', '/bookmarks'),
  followTopic: (topicId: any) => req('POST', `/topics/${topicId}/follow`),

  // users
  getUser: (username: any) => req('GET', `/users/${username}`),
  updateProfile: (r: any) => req('PUT', '/user/me', r),
  getUserTopics: (username: any, page?: any) =>
    req('GET', `/users/${username}/topics${page ? '?page=' + page : ''}`),
  getUserPosts: (username: any, page?: any) =>
    req('GET', `/users/${username}/posts${page ? '?page=' + page : ''}`),

  // admin
  listUsers: (params: any = {}) => {
    const q = new URLSearchParams();
    if (params.q) q.set('q', params.q);
    if (params.page) q.set('page', params.page);
    return req('GET', `/admin/users?${q.toString()}`);
  },
  adminUpdateUser: (username: any, r: any) => req('PUT', `/admin/users/${username}`, r),
  listReports: (status?: any) => req('GET', `/reports${status ? '?status=' + status : ''}`),
  resolveReport: (id: any) => req('POST', `/reports/${id}/resolve`),
  dismissReport: (id: any) => req('POST', `/reports/${id}/dismiss`),
};
