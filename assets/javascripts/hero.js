/**
 * 全屏吸引子动效 —— Zensical 官网首页技法的精简实现
 *
 * 用法：页面上放一个带 data-hero 属性的容器即可，例如
 *   <div class="home-hero" data-hero></div>
 *
 * 技术要点（与 zensical.org 一致）：
 *   1. 全屏 WebGL canvas 作为背景层
 *   2. Three.js + EffectComposer：RenderPass + AfterimagePass（残影通道制造拖尾辉光）
 *   3. OrbitControls：鼠标拖拽旋转 / 滚轮缩放
 *   4. 粒子沿「奇怪吸引子」轨迹流动，残影累积成缠绕的光轨
 *   5. 轨迹存在环形缓冲里持续重积分，且吸引子参数 b 缓慢漂移 ——
 *      曲线本身会随时间自行变形，而不是一条固定轨迹
 *   6. 点击「向下滚动」提示可平滑滚到首屏内容
 *   7. WebGL 不可用时给 <html> 加 .no-webgl，由 CSS 降级为静态渐变
 *   8. 容器高度自适应视口（可用 data-hero-offset 指定要减去的固定高度，如顶栏）
 *   9. 生命周期跟随 Material 的 document$：每次导航先卸载旧实例，再按新 DOM 挂载
 *
 * 可调参数集中在 CONFIG 里。
 *
 * 部署：overrides/main.html 引用的是打包产物 hero.bundle.js（同源自托管，
 * 避免原生 ESM 经 CDN 按依赖逐层拉取 ~10 个文件）。改动本文件或升级
 * three 后需重新打包：make hero-bundle（node 工程在 scripts/hero/ 下），
 * 并把新的 hero.bundle.js 提交进仓库。
 */
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { EffectComposer } from 'three/addons/postprocessing/EffectComposer.js';
import { RenderPass } from 'three/addons/postprocessing/RenderPass.js';
import { AfterimagePass } from 'three/addons/postprocessing/AfterimagePass.js';

const CONFIG = {
  steps: 300000,      // 环形轨迹缓冲长度（越大曲线越密）
  dt: 0.002,         // 积分步长：更小 → 折线更贴合真实曲线（细腻度主参数）
  b: 0.18,            // 吸引子参数基准值：0.19 最稳定，0.13~0.21 形态各异
  bSwing: 0.045,      // b 的自动漂移幅度：形态随时间变化的来源（覆盖 0.135~0.225）
  bPeriod: 40,        // 漂移周期（秒）
  rebuildSeconds: 1.2, // 环形缓冲整体换一遍所需时间（秒）。必须远小于 bPeriod：
                       // 否则缓冲里会叠着好几个 b 的形态，曲线会「糊」成一团
  count: 6000,        // 粒子数：铺得更满，流光更连贯
  speed: 100,         // 粒子沿轨迹的视觉流速（索引/帧）；速度 ∝ speed × dt，
                      // dt 调小后等比上调以保持原有流速
  size: 0.005,        // 粒子尺寸：更小的点颗粒更细，靠叠加辉光补亮度
  damp: 0.90,         // 残影衰减：越大拖尾越长越柔；>0.95 容易糊成一片
  rescanEvery: 30,    // 每隔多少帧重新量一次缓冲尺度
  scaleLerp: 0.06,    // 尺度趋近速度：越小越平滑
  color: 0xff7735,    // 主色（官网同款橙）
  bg: 0x0b0c0f,       // 背景色
  maxPixelRatio: 2,   // 像素比上限（性能兜底）
  viewAzimuth: -109.7, // 初始水平方位角（度）：0 = 正面，正值绕竖直轴向右转
  viewElevation: -32.8, // 初始仰角（度）：0 = 平视，正值从上往下俯视（负值从下往上仰视）
  debugHud: true,    // 调参 HUD：左下角实时显示相机机位与等效参数，需要再调时改 true
};

/* ---------- 挂载 / 卸载 ----------
   本站启用了 navigation.instant：Material 会就地替换
   [data-md-component=container]，hero 容器正好在其中；同时 ES 模块按 URL
   只求值一次，导航回首页时本文件不会重新执行。若在顶层直接 init，
   第二次回到首页就只剩空壳 DOM（黑底 + 文字），且旧的渲染循环仍在后台空转。

   因此改为订阅 Material 暴露的 document$（bundle 中的 ReplaySubject：
   订阅即回放当前文档，之后每次导航都会推送一次），由它驱动
   「先卸载旧实例，再按新 DOM 挂载」。模块只求值一次恰好保证
   全程只有一个订阅，不会重复叠加。 */
let instance = null;

function mount() {
  const host = document.querySelector('[data-hero]');

  // 同一容器被重复触发：无需重建
  if (instance && instance.host === host) return;

  // 容器已随导航被替换：释放上一轮的渲染循环与 WebGL 资源
  if (instance) {
    instance.dispose();
    instance = null;
  }

  if (!host) return;

  if (!document.createElement('canvas').getContext('webgl')) {
    document.documentElement.classList.add('no-webgl');
    return;
  }

  instance = init(host);
}

if (window.document$ && typeof window.document$.subscribe === 'function') {
  window.document$.subscribe(mount);
} else {
  // 兜底：bundle 未就绪，或未来版本不再暴露 document$，退化为一次性初始化
  mount();
}

function init(host) {
  /* ---------- 高度自适应：撑满首屏 ----------
     预留高度优先取 data-hero-offset；未指定时自动取顶栏高度。
     注意：本站启用了 navigation.tabs.sticky，tabs 位于 .md-header 内部，
     因此 header 的 offsetHeight 已包含 tabs，无需重复相减。 */
  const header = document.querySelector('.md-header');

  function reserveHeight() {
    if (host.dataset.heroOffset !== undefined) {
      return Number(host.dataset.heroOffset) || 0;
    }
    return header ? header.offsetHeight : 0;
  }

  function fitHeight() {
    host.style.height = `${Math.max(320, window.innerHeight - reserveHeight())}px`;
  }
  fitHeight();

  /* ---------- 渲染器 ---------- */
  const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, CONFIG.maxPixelRatio));
  host.appendChild(renderer.domElement);

  /* ---------- 场景与相机 ---------- */
  const scene = new THREE.Scene();
  scene.background = new THREE.Color(CONFIG.bg);

  const camera = new THREE.PerspectiveCamera(60, 1, 0.01, 100);

  /* ---------- 初始机位：距离自适应，角度来自 CONFIG ----------
     吸引子经 rescanScale 归一化后整体近似落在半径 ~1.2 的球内；
     FIT_RADIUS 是相对该球的比例，越小相机越近、画面越满，
     当前值来自 HUD 实测的最佳视角。fitDistance 按「球面顶到视口
     边缘」反推距离：宽屏以视高为准（两侧允许少量溢出，更沉浸），
     窄屏同时约束视宽防溢出。机位只在挂载时算一次，用户手动
     拖拽/缩放后 resize 不会重置。 */
  const FIT_RADIUS = 0.609;

  function fitDistance() {
    const half = (camera.fov * Math.PI) / 360; // 垂直半视角
    const aspect = (host.clientWidth || 1) / (host.clientHeight || 1);
    const distV = FIT_RADIUS / Math.tan(half);
    const distH = FIT_RADIUS / (Math.tan(half) * aspect);
    return aspect >= 1 ? distV : Math.max(distV, distH);
  }
  // 角度：相机放在半径 fitDist 的球面上，由 CONFIG 的方位角/仰角定位，
  // OrbitControls 会让它始终看向原点（距离自适应逻辑见上方 fitDistance）
  const azimuth = (CONFIG.viewAzimuth * Math.PI) / 180;
  const elevation = (CONFIG.viewElevation * Math.PI) / 180;
  const fitDist = fitDistance();
  camera.position.set(
    fitDist * Math.cos(elevation) * Math.sin(azimuth),
    fitDist * Math.sin(elevation),
    fitDist * Math.cos(elevation) * Math.cos(azimuth)
  );

  /* ---------- 吸引子轨迹：环形缓冲 + 持续重积分 ----------
     Thomas 循环对称吸引子：
       dx/dt = sin(y) - b·x
       dy/dt = sin(z) - b·y
       dz/dt = sin(x) - b·z

     轨迹不是一次性算完的：每帧都接着往下积分 growth 步并写进环形缓冲，
     同时让 b 沿正弦缓慢漂移。因为 ODE 是耗散的，轨迹会一直贴着「当前 b
     对应的吸引子」走，于是整条曲线会随时间自己变形，而不是固定不动。 */
  const traj = new Float32Array(CONFIG.steps * 3);
  let cx = 0.1;
  let cy = 0;
  let cz = 0;
  let cursor = 0; // 写指针，指向缓冲里最旧的一个点

  // b 随时间在 [b - bSwing, b + bSwing] 之间往复
  function attractorB(seconds) {
    return CONFIG.b + CONFIG.bSwing * Math.sin((seconds / CONFIG.bPeriod) * Math.PI * 2);
  }

  function integrate(b) {
    const dx = Math.sin(cy) - b * cx;
    const dy = Math.sin(cz) - b * cy;
    const dz = Math.sin(cx) - b * cz;
    cx += dx * CONFIG.dt;
    cy += dy * CONFIG.dt;
    cz += dz * CONFIG.dt;
    traj[cursor * 3] = cx;
    traj[cursor * 3 + 1] = cy;
    traj[cursor * 3 + 2] = cz;
    cursor = cursor + 1 === CONFIG.steps ? 0 : cursor + 1;
  }

  // 先把整条缓冲填满，避免开场是一段空白
  for (let i = 0; i < CONFIG.steps; i++) integrate(attractorB(0));

  /* ---------- 粒子 ---------- */
  const positions = new Float32Array(CONFIG.count * 3);
  const head = new Float32Array(CONFIG.count);
  for (let i = 0; i < CONFIG.count; i++) {
    // 沿轨迹均匀铺开
    head[i] = (i / CONFIG.count) * CONFIG.steps;
  }

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));

  const material = new THREE.PointsMaterial({
    color: CONFIG.color,
    size: CONFIG.size,
    sizeAttenuation: true,
    transparent: true,
    opacity: 0.95,
    blending: THREE.AdditiveBlending,
    depthWrite: false,
  });
  const points = new THREE.Points(geometry, material);
  scene.add(points);

  /* ---------- 自适应缩放 ----------
     b 漂移会让吸引子的整体尺度变化，所以缓冲里存原始坐标，缩放交给
     points.scale。周期性重新量一次缓冲里的最大坐标，再让系数平滑趋近，
     避免整条曲线在重算的瞬间跳一下。 */
  let scale = 1;
  let scaleTarget = 1;
  let framesToRescan = 0;

  function rescanScale() {
    let maxAbs = 0;
    for (let i = 0; i < traj.length; i++) {
      const v = Math.abs(traj[i]);
      if (v > maxAbs) maxAbs = v;
    }
    scaleTarget = 1 / (maxAbs || 1);
  }

  rescanScale();
  scale = scaleTarget;
  points.scale.setScalar(scale);

  /* ---------- 后处理：AfterimagePass 是拖尾/辉光的来源 ---------- */
  const afterimage = new AfterimagePass(CONFIG.damp);
  const composer = new EffectComposer(renderer);
  composer.addPass(new RenderPass(scene, camera));
  composer.addPass(afterimage);

  /* ---------- 交互 ---------- */
  const controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.dampingFactor = 0.08;
  controls.rotateSpeed = 1.2;
  controls.enablePan = false;
  controls.minDistance = 0.6;
  controls.maxDistance = 6;

  /* ---------- 调参 HUD：拖拽/缩放时实时打印相机机位 ----------
     把相机位置换算回球坐标：方位角/仰角可直接抄进 CONFIG.viewAzimuth /
     viewElevation；缩放距离与 FIT_RADIUS 成正比（dist ∝ FIT_RADIUS），
     据此给出等效 FIT_RADIUS。拖拽结束时会往控制台打一条同样的日志，
     方便直接复制。调完后把 CONFIG.debugHud 改为 false 整体关闭。 */
  let hud = null;
  let refreshHud = null;
  let logHud = null;

  if (CONFIG.debugHud) {
    hud = document.createElement('pre');
    hud.style.cssText =
      'position:absolute;left:12px;bottom:1.4rem;z-index:2;margin:0;' +
      'padding:6px 10px;font:12px/1.7 ui-monospace,SFMono-Regular,monospace;' +
      'color:rgba(255,255,255,.78);background:rgba(0,0,0,.35);' +
      'border-radius:6px;pointer-events:none;white-space:pre;';
    host.appendChild(hud);

    function hudText() {
      const p = camera.position;
      const r = p.length();
      const az = (Math.atan2(p.x, p.z) * 180) / Math.PI;
      const el = (Math.asin(p.y / r) * 180) / Math.PI;
      return (
        `x ${p.x.toFixed(3)}  y ${p.y.toFixed(3)}  z ${p.z.toFixed(3)}\n` +
        `az ${az.toFixed(1)}  el ${el.toFixed(1)}`
        // `FIT_RADIUS: ${(FIT_RADIUS * (r / fitDist)).toFixed(3)}`
      );
    }
    refreshHud = () => { hud.textContent = hudText(); };
    logHud = () => { console.log('[hero 机位]', hudText()); };
    controls.addEventListener('change', refreshHud);
    controls.addEventListener('end', logHud);
    refreshHud();
  }

  /* ---------- 滚动提示：点击滚到首屏内容 ----------
     减掉顶栏高度，让正文正好落在吸顶的顶栏下方。 */
  const scrollHint = host.querySelector('.home-hero__scroll');

  function scrollToContent() {
    const main = document.querySelector('.md-main');
    const top = main
      ? main.getBoundingClientRect().top + window.scrollY - (header ? header.offsetHeight : 0)
      : host.offsetHeight;
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    window.scrollTo({ top, behavior: reduced ? 'auto' : 'smooth' });
  }

  if (scrollHint) scrollHint.addEventListener('click', scrollToContent);

  /* ---------- 尺寸自适应 ---------- */
  function resize() {
    fitHeight();
    const w = host.clientWidth || 1;
    const h = host.clientHeight || 1;
    camera.aspect = w / h;
    camera.updateProjectionMatrix();
    renderer.setSize(w, h);
    composer.setSize(w, h);
  }
  window.addEventListener('resize', resize);
  resize();

  /* ---------- 主循环 ---------- */
  let rafId = 0;
  const startedAt = performance.now();
  let lastAt = startedAt;

  function tick() {
    const now = performance.now();
    const seconds = (now - startedAt) / 1000;
    const dtSec = Math.min(0.25, (now - lastAt) / 1000);
    lastAt = now;

    // 积分按时间推进而不是按帧：慢机器上曲线不会因为「来不及重画」
    // 而把好几个形态叠在一起
    const growth = Math.round((CONFIG.steps / CONFIG.rebuildSeconds) * dtSec);

    // 曲线自我重塑：接着往下积分，同时把 b 缓慢推向下一形态
    const b = attractorB(seconds);
    for (let i = 0; i < growth; i++) integrate(b);

    // 缓冲自己往前滑了 growth 步，粒子的逻辑索引要反向补掉这一段，
    // 世界位置才会以 CONFIG.speed 的视觉速度前进
    const headAdvance = CONFIG.speed - growth;

    if (--framesToRescan <= 0) {
      rescanScale();
      framesToRescan = CONFIG.rescanEvery;
    }
    if (Math.abs(scaleTarget - scale) > 1e-5) {
      scale += (scaleTarget - scale) * CONFIG.scaleLerp;
      points.scale.setScalar(scale);
    }

    for (let i = 0; i < CONFIG.count; i++) {
      let h = head[i] + headAdvance;
      if (h >= CONFIG.steps) h -= CONFIG.steps;
      else if (h < 0) h += CONFIG.steps;
      head[i] = h;

      let p = cursor + (h | 0);
      if (p >= CONFIG.steps) p -= CONFIG.steps;
      const s = p * 3;
      positions[i * 3] = traj[s];
      positions[i * 3 + 1] = traj[s + 1];
      positions[i * 3 + 2] = traj[s + 2];
    }
    geometry.attributes.position.needsUpdate = true;
    controls.update();
    composer.render();
    rafId = requestAnimationFrame(tick);
  }
  tick();

  /* ---------- 卸载 ----------
     导航离开首页时由 mount() 调用：停掉渲染循环、摘掉监听器、
     释放几何/材质/后处理与 WebGL 上下文。浏览器同时存活的 WebGL
     上下文数量有限，不主动释放会在多次往返后触发上下文丢失。 */
  let disposed = false;

  function dispose() {
    if (disposed) return;
    disposed = true;

    cancelAnimationFrame(rafId);
    window.removeEventListener('resize', resize);
    if (scrollHint) scrollHint.removeEventListener('click', scrollToContent);
    renderer.domElement.remove();
    if (hud) {
      controls.removeEventListener('change', refreshHud);
      controls.removeEventListener('end', logHud);
      hud.remove();
    }

    controls.dispose();
    geometry.dispose();
    material.dispose();
    // EffectComposer.dispose() 只回收自身的 render target，
    // pass 内部的目标要各自释放
    afterimage.dispose();
    composer.dispose();
    renderer.dispose();
    renderer.forceContextLoss();
  }

  return { host, dispose };
}
