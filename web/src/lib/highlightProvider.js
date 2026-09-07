// 按需代码高亮的分层加载器。
// - highlightBundle 静态归集 refractor + 语言并注册；
// - 本文件只用动态 import() 访问它，把 refractor 隔离到独立异步 chunk，
//   主包初始化不引用 refractor，从而避免生成非法初始化顺序（TDZ）。

let _refractor = null; // 已就绪的高亮器（同步可用）
let _promise = null;   // 加载中的 promise（去重）

/** 返回已就绪的 refractor；未就绪返回 null（本轮渲染跳过高亮）。 */
export function getRefractor() {
  return _refractor;
}

/** 懒加载高亮器（幂等）。页面挂载时调用即可，之后同步高亮可用。 */
export function ensureSyntaxHighlight() {
  if (!_promise) {
    _promise = import('./highlightBundle')
      .then((m) => {
        _refractor = m.refractor;
        return _refractor;
      })
      .catch((err) => {
        _promise = null; // 失败后允许重试
        throw err;
      });
  }
  return _promise;
}