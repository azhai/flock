import m from 'mithril';
import { api } from '../lib/api';
import { store } from '../lib/store';
import { Layout } from '../components/Layout';
import { Button, toast } from '../lib/ui';
import { A } from '../lib/route';

export const Register: m.Component = {
  oninit() {
    if (store.user) m.route.set('/');
  },
  view() {
    return m(Layout, [
      m('.center-screen', [
        m('.auth-card', [
          m('div', { class: 'text-center mb-4' }, [
            m('div', { class: 'site-logo', style: { justifyContent: 'center', fontSize: '20px' } }, [m('div', { class: 'mark', style: { width: '34px', height: '34px', fontSize: '18px' } }, 'F'), 'Flock']),
            m('h2', { class: 'mt-3', style: { margin: '8px 0 0' } }, '注册'),
          ]),
          m('form', {
            onsubmit: (e: any) => {
              e.preventDefault();
              const fd = new FormData(e.target);
              const username = String(fd.get('username') || '').trim();
              const email = String(fd.get('email') || '').trim();
              const password = String(fd.get('password') || '');
              const confirm = String(fd.get('confirm') || '');
              if (!username || !email || !password) return toast('请填写完整信息', 'warning');
              if (password !== confirm) return toast('两次密码不一致', 'warning');
              api
                .register({ username, email, password })
                .then((d: any) => {
                  store.setUser(d.user, d.token);
                  toast('注册成功', 'success');
                  m.route.set('/');
                })
                .catch((err) => toast(err || '注册失败', 'error'));
            },
          }, [
            m('label', { class: 'field-label' }, '用户名'),
            m('input', { class: 'input', name: 'username', placeholder: '用户名' }),
            m('label', { class: 'field-label mt-3' }, '邮箱'),
            m('input', { class: 'input', name: 'email', type: 'email', placeholder: '邮箱' }),
            m('label', { class: 'field-label mt-3' }, '密码'),
            m('input', { class: 'input', name: 'password', type: 'password', placeholder: '至少 6 位' }),
            m('label', { class: 'field-label mt-3' }, '确认密码'),
            m('input', { class: 'input', name: 'confirm', type: 'password', placeholder: '再次输入密码' }),
            m(Button, { primary: true, label: '注册', type: 'submit', class: 'full mt-4' }),
          ]),
          m('.text-center.mt-3.text-sm muted', [A('/login', {}, '已有账号？去登录')]),
        ]),
      ]),
    ]);
  },
};
