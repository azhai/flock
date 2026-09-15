import m from 'mithril';

const KEY_U = 'user';
const KEY_T = 'token';

function loadUser(): any {
  try {
    return JSON.parse(localStorage.getItem(KEY_U) || 'null');
  } catch {
    return null;
  }
}

export const store = {
  user: loadUser(),
  token: localStorage.getItem(KEY_T) || null,
  headerTitle: null as string | null,

  setUser(user: any, token: string) {
    this.user = user;
    this.token = token;
    localStorage.setItem(KEY_U, JSON.stringify(user));
    localStorage.setItem(KEY_T, token);
    m.redraw();
  },

  setHeaderTitle(t: string | null) {
    this.headerTitle = t;
    m.redraw();
  },

  logout() {
    this.user = null;
    this.token = null;
    this.headerTitle = null;
    localStorage.removeItem(KEY_U);
    localStorage.removeItem(KEY_T);
    m.redraw();
  },
};
