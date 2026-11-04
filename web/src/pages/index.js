import { home } from './home.js';
import { topic, postRow } from './topic.js';
import { newTopic } from './new-topic.js';
import { login, register } from './auth.js';
import { bookmarks, profile } from './misc.js';
import { admin } from './admin.js';

// 页面注册表（capability seam）：新增页面 = 注册一个 Alpine.data + 一个模板。
export function registerPages(Alpine) {
  Alpine.data('home', home);
  Alpine.data('topic', topic);
  Alpine.data('postRow', postRow);
  Alpine.data('newTopic', newTopic);
  Alpine.data('login', login);
  Alpine.data('register', register);
  Alpine.data('bookmarks', bookmarks);
  Alpine.data('profile', profile);
  Alpine.data('admin', admin);
}
