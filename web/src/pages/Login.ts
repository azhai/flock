import m from 'mithril';
import { api } from '../lib/api';
import { store } from '../lib/store';
import { Layout } from '../components/Layout';
import { Button, toast } from '../lib/ui';
import { A } from '../lib/route';

export const Login: m.Component = {
  oninit() {
    if (store.user) m.route.set('/');
  },
  view() {
    const st: any = {};
    return m(Layout, [
      m('.center-screen', [
        m('.auth-card', [
          m('div', { class: 'text-center mb-4' }, [
            m('div', { class: 'site-logo', style: { justifyContent: 'center', fontSize: '20px' } }, [m('div', { class: 'mark', style: { width: '34px', height: '34px', fontSize: '18px' } }, 'F'), 'Flock']),
            m('h2', { class: 'mt-3', style: { margin: '8px 0 0' } }, '登录'),
          ]),
          m('form', {
            onsubmit: (e: any) => {
              e.preventDefault();
              const fd = new FormData(e.target);
              const username = String(fd.get('username') || '').trim();
              const password = String(fd.get('password') || '');
              if (!username || !password) return toast('请输入用户名和密码', 'warning');
              api
                .login({ username, password })
                .then((d: any) => {
                  store.setUser(d.user, d.token);
                  toast('登录成功', 'success');
                  m.route.set('/');
                })
                .catch((err) => toast(err || '登录失败', 'error'));
            },
          }, [
            m('label', { class: 'field-label' }, '用户名或邮箱'),
            m('input', { class: 'input', name: 'username', placeholder: '用户名或邮箱', autocomplete: 'username' }),
            m('label', { class: 'field-label mt-3' }, '密码'),
            m('input', { class: 'input', name: 'password', type: 'password', placeholder: '密码', autocomplete: 'current-password' }),
            m(Button, { primary: true, label: '登录', type: 'submit', class: 'full mt-4' }),
          ]),
          m('.text-center.mt-3.text-sm muted', [A('/register', { class: '' }, '还没有账号？立即注册')]),
        ]),
      ]),
    ]);
  },
};
