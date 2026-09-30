/* ===========================================================================
   脚注 tooltip 滚动滞留补丁（Zensical 0.1.0 / bundle.b7d3f789）
   ===========================================================================
   现象：悬停脚注引用后滚动页面，卡片不消失——主题的 active 状态只由
   mouseenter/mouseleave 与"悬停卡片本体"驱动（Wr 函数内的 RxJS 状态机），
   视口滚动既不触发 mouseleave 也不参与取消激活。旧卡片滞留 --active，
   再悬停另一个引用时出现两张卡片同屏。

   补丁：视口滚动时收起所有激活的 tooltip。只移除展示用 class，不触碰
   主题的状态机——下一次 mouseenter/focus 会正常重新激活。

   - passive 监听 + rAF 节流，滚动零额外开销
   - window.__tooltipScrollFix 幂等标记：navigation.instant 下脚本被重复
     执行时监听器只挂一次
   =========================================================================== */

(function() {
  if (window.__tooltipScrollFix) return;
  window.__tooltipScrollFix = true;

  var ticking = false;

  window.addEventListener("scroll", function() {
    if (ticking) return;
    ticking = true;
    requestAnimationFrame(function() {
      document.querySelectorAll(".md-tooltip2--active").forEach(function(el) {
        el.classList.remove("md-tooltip2--active");
      });
      ticking = false;
    });
  }, { passive: true });
})();
