import m from 'mithril';
import { Header } from './Header';
import { ToastHost } from '../lib/ui';

export const Layout: m.Component = {
  view(vnode) {
    return [m(Header), ToastHost(), m('.container', vnode.children)];
  },
};
