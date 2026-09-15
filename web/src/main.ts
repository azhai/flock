import m from 'mithril';
import 'construct-ui/lib/index.css';
import './styles.css';
import './markdown.css';
import { Layout } from './components/Layout';
import { Home } from './pages/Home';
import { TopicDetail } from './pages/TopicDetail';
import { NewTopic } from './pages/NewTopic';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { UserProfile } from './pages/UserProfile';
import { Bookmarks } from './pages/Bookmarks';
import { AdminPanel } from './pages/AdminPanel';

const NotFound = {
  view: () =>
    m(Layout, [
      m('.center-screen', [
        m('.empty', [
          m('div', { class: 'big' }, '404'),
          m('div', {}, '页面不存在'),
          m('button', { class: 'btn btn-primary mt-3', onclick: () => m.route.set('/') }, '返回首页'),
        ]),
      ]),
    ]),
};

const root = document.getElementById('root') as HTMLElement;
m.route(root, '/', {
  '/': Home,
  '/topics/:id': TopicDetail,
  '/new': NewTopic,
  '/login': Login,
  '/register': Register,
  '/u/:username': UserProfile,
  '/bookmarks': Bookmarks,
  '/admin': AdminPanel,
  '/*': NotFound,
});
