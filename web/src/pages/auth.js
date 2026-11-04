import { api } from '../core/api.js';

// from 来自 URL，只接受站内绝对路径，避免开放重定向。
function safeFrom(raw) {
  const v = String(raw || '/');
  return v.startsWith('/') && !v.startsWith('//') && !v.startsWith('/\\') ? v : '/';
}

export function login() {
  return {
    username: '',
    password: '',
    submitting: false,
    get from() { return safeFrom(this.$store.route.query.from); },
    async submit() {
      if (!this.username.trim() || !this.password.trim()) {
        this.$store.ui.warn('请填写用户名和密码');
        return;
      }
      this.submitting = true;
      try {
        const d = await api.login({ username: this.username.trim(), password: this.password });
        this.$store.session.setSession(d.user, d.token);
        this.$store.ui.ok('登录成功');
        this.$go(this.from);
      } catch {
        this.$store.ui.fail('用户名或密码错误');
      } finally {
        this.submitting = false;
      }
    },
  };
}

export function register() {
  return {
    username: '',
    email: '',
    password: '',
    submitting: false,
    get from() { return safeFrom(this.$store.route.query.from); },
    async submit() {
      if (!this.username.trim() || !this.email.trim() || !this.password.trim()) {
        this.$store.ui.warn('请填写所有字段');
        return;
      }
      if (this.password.length < 6) {
        this.$store.ui.warn('密码至少 6 位');
        return;
      }
      this.submitting = true;
      try {
        const d = await api.register({
          username: this.username.trim(), email: this.email.trim(), password: this.password,
        });
        this.$store.session.setSession(d.user, d.token);
        this.$store.ui.ok('注册成功');
        this.$go(this.from);
      } catch (e) {
        this.$store.ui.fail(e || '注册失败');
      } finally {
        this.submitting = false;
      }
    },
  };
}
