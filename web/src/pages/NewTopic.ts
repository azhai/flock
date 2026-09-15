import m from 'mithril';
import { api } from '../lib/api';
import { store } from '../lib/store';
import { Layout } from '../components/Layout';
import { Button, toast } from '../lib/ui';
import { TagInput } from '../components/TagInput';
import { MarkdownView } from '../components/MarkdownView';
import { Icon } from '../lib/icon';

export const NewTopic: m.Component = {
  oninit(vnode: any) {
    const st = vnode.state as any;
    st.title = '';
    st.content = '';
    st.categoryId = '';
    st.tags = [];
    st.cats = [];
    st.preview = false;
    st.loading = false;
    api.listCategories().then((c: any) => {
      st.cats = c || [];
      if (st.cats.length) st.categoryId = st.cats[0].id;
      m.redraw();
    }).catch(() => {});
  },
  view(vnode: any) {
    if (!store.user) {
      m.route.set('/login');
      return m('div');
    }
    const st = vnode.state;
    return m(Layout, [
      m('.narrow', [
        m('h1', { class: 'text-xl fw-700 mb-4' }, '发布新话题'),
        m('label', { class: 'field-label' }, '标题'),
        m('input', { class: 'input', value: st.title, oninput: (e: any) => (st.title = e.target.value), placeholder: '一句话说明你的问题或想法' }),
        m('label', { class: 'field-label mt-3' }, '分类'),
        m('select', { class: 'select', value: st.categoryId, oninput: (e: any) => (st.categoryId = e.target.value) },
          st.cats.map((c: any) => m('option', { value: c.id }, c.name))
        ),
        m('label', { class: 'field-label mt-3' }, '标签'),
        m(TagInput, { value: st.tags, onchange: (v: any) => (st.tags = v) }),
        m('label', { class: 'field-label mt-3' }, '正文（支持 Markdown）'),
        m('.editor-wrap', [
          m('.editor-toolbar', [
            Button({ size: 'xs', ghost: true, label: st.preview ? '编辑' : '预览', icon: st.preview ? 'edit' : 'eye', onclick: () => (st.preview = !st.preview) }),
          ]),
          st.preview ? m('.md-preview', [MarkdownView({ src: st.content })]) : m('textarea', { value: st.content, oninput: (e: any) => (st.content = e.target.value), placeholder: '用 Markdown 写下正文…', style: { minHeight: '220px' } }),
        ]),
        m('.btn-row.mt-4', [
          Button({ primary: true, label: '发布', icon: 'send', disabled: st.loading, onclick: () => {
            if (!st.title.trim()) return toast('请填写标题', 'warning');
            if (!st.content.trim()) return toast('请填写正文', 'warning');
            st.loading = true;
            api.createTopic({ title: st.title.trim(), content: st.content, category_id: st.categoryId, tags: st.tags })
              .then((d: any) => { toast('发布成功', 'success'); m.route.set('/topics/' + d.id); })
              .catch((err) => { st.loading = false; toast(err || '发布失败', 'error'); });
          } }),
          Button({ ghost: true, label: '取消', onclick: () => m.route.set('/') }),
        ]),
      ]),
    ]);
  },
};
