# progress.md — 去构建链（本地 vendor min 版）

| 时间 | 动作 | 结果 |
| --- | --- | --- |
| 2026-11-05 05:05 | 用户要求：极简去掉 pnpm/Vite，静态文件下载最新 min 版方便 embed | 收到 |
| 2026-11-05 05:08 | 读 main.go / main.js / markdown.js / index.html / Makefile / package.json；Glob views | 确认 embed web/dist、?raw、裸 alpinejs、SPA fallback 已支持任意静态文件 |
| 2026-11-05 05:10 | 查 web 根 + 探测 CDN（curl 全 000，禁外网）→ 改用 node_modules 复制 | 确认 node_modules/dist/public 存在 |
| 2026-11-05 05:12 | 确认依赖文件形态 | alpine module.esm.min.js(ESM)、marked.umd.js(UMD)、purify.min.js(UMD min)、microlight.js(UMD)；public 含 2 svg |
| 2026-11-05 05:14 | 写 .planning/2026-11-05-cdn-vendor/{task_plan,findings,progress}.md | 成功 |
| 2026-11-05 05:22 | 复制 vendor(4) + svg(2) 到 web/ | 成功 |
| 2026-11-05 05:23 | 改 main.go(embed web)、main.js(fetch 模板+去 css import)、markdown.js(全局)、index.html(importmap+UMD+link)、Makefile(去 pnpm/vite) | 成功 |
| 2026-11-05 05:24 | 删 node_modules/dist/public/package/pnpm/vite/DS_Store；go build + go test | build 成功(25M)；test 全绿 |
| 2026-11-05 05:29 | 启动 /tmp/flock-cdn @8093，HTTP 冒烟 | health 200；所有静态资源 + SPA fallback 200 |
| 2026-11-05 05:33 | vendor 文件格式/符号校验（grep + node import） | alpine ESM default export；marked/dompurify/microlight UMD 全局，版本正确 |
| 2026-11-05 05:35 | 停验证服务(7iip9m)；更新 README/ADR 0004/planning/memory | 完成 |

## 最终验证结论
- 构建：`go build` 通过（embed `web`，25M 二进制）。
- 测试：`go test ./...` 全绿；go.mod/go.sum 用备份+trap 恢复，无残留。
- 运行时（HTTP 层）：health、所有 `/src/*`、`/vendor/*`、`/src/views/*.html`、CSS、SVG、SPA fallback 均 200。
- vendor 库：格式与导出符号校验通过，版本与 package.json 锁定一致。
- 限制：沙箱无 chromium / 无全局 puppeteer，未能做无头浏览器逐页回归；建议本地 `make run` 后用浏览器实测首页/话题/发帖/高亮/通知。

## 已删除（生成物/依赖，非源码）
web/node_modules、web/dist、web/public、web/package.json、web/pnpm-lock.yaml、web/pnpm-workspace.yaml、web/vite.config.js、web/.DS_Store

## 新增
web/vendor/{alpine.module.esm.min.js, marked.umd.js, purify.min.js, microlight.js}
web/{favicon.svg, logo-forest.svg}
