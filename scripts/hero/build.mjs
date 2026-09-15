/**
 * 打包首页 hero 动效：docs/assets/javascripts/hero.js → hero.bundle.js
 *
 * 为什么需要打包：
 *   原生 ESM + importmap 走 CDN 时，浏览器要按依赖链逐层发现并拉取
 *   three 本体 + 5 个 addons（约 10 个文件、3 层串行 RTT），首屏等待明显；
 *   jsdelivr 在国内也不稳定。打成单文件后同源自托管，只有 1 个请求，
 *   可被浏览器长期缓存，且不再依赖第三方 CDN。
 *
 * 用法：make hero-bundle
 *   （即：cd scripts/hero && npm install && npm run build:hero）
 * 注意：three 版本固定在 package.json（0.169.0，与历史 importmap 一致）；
 *       升级 three 后改版本号并重新执行本脚本，产物需提交仓库。
 */
import { build } from 'esbuild';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

// 以本文件为基准解析路径，无论从哪个工作目录调用都指向同一个源文件与产物
const here = dirname(fileURLToPath(import.meta.url));
const docs = resolve(here, '../../docs/assets/javascripts');

await build({
  entryPoints: [resolve(docs, 'hero.js')],
  outfile: resolve(docs, 'hero.bundle.js'),
  // hero.js 在 docs/ 下，不在本 node 工程目录树内。Node 解析裸导入只会从
  // hero.js 的祖先目录逐级向上找 node_modules（docs/ → 根，均无），够不着
  // scripts/hero/node_modules，因此需用 nodePaths（等价 NODE_PATH）显式追加
  // 包搜索目录，否则报 "Could not resolve three"。
  nodePaths: [resolve(here, 'node_modules')],
  bundle: true,
  minify: true,
  format: 'esm',
  target: ['es2020'],
  legalComments: 'none',
  logLevel: 'info',
});
