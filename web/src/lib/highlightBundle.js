// 语法高亮 bundle：集中静态引入 refractor + 白名单语言并完成注册。
// 该模块只能被 highlightProvider 动态 import() 访问，因此主包初始化不引用
// refractor，从根本上避免 Rollup 因 refractor 循环依赖生成非法初始化顺序
// （"Cannot access 'X' before initialization"）。
import { refractor } from 'refractor/core';
import clike from 'refractor/clike';
import c from 'refractor/c';
import cpp from 'refractor/cpp';
import java from 'refractor/java';
import css from 'refractor/css';
import markup from 'refractor/markup';
import javascript from 'refractor/javascript';
import jsx from 'refractor/jsx';
import typescript from 'refractor/typescript';
import tsx from 'refractor/tsx';
import json from 'refractor/json';
import go from 'refractor/go';
import python from 'refractor/python';
import bash from 'refractor/bash';
import markdown from 'refractor/markdown';
import sql from 'refractor/sql';
import yaml from 'refractor/yaml';
import csharp from 'refractor/csharp';
import rust from 'refractor/rust';
import php from 'refractor/php';
import ruby from 'refractor/ruby';
import swift from 'refractor/swift';
import kotlin from 'refractor/kotlin';
import scss from 'refractor/scss';
import less from 'refractor/less';

// 按依赖顺序注册（base 先于依赖者）
for (const lang of [
  clike, c, cpp, java, css, markup, javascript, jsx, typescript, tsx, json,
  go, python, bash, markdown, sql, yaml, csharp, rust, php, ruby, swift,
  kotlin, scss, less,
]) refractor.register(lang);

export { refractor };