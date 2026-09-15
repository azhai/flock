import m from 'mithril';
import { store } from './store';

/** 站内导航：阻止默认跳转并用 m.route 接管 */
export function go(path: string) {
  store.setHeaderTitle(null);
  m.route.set(path);
}

export function A(path: string, opts: any, children: any) {
  return m(
    'a',
    Object.assign(
      {
        href: path,
        onclick: (e: any) => {
          e.preventDefault();
          store.setHeaderTitle(null);
          m.route.set(path);
        },
      },
      opts || {}
    ),
    children
  );
}
