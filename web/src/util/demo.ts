/**
 * 演示模式判定工具
 * 优先级覆盖：
 * 1. 运行时动态标记: setRuntimeDemoMode(true)
 * 2. 构建期/静态环境变量: import.meta.env.VITE_RUN_MODE === 'demo'
 * 3. 域名特征: window.location.hostname 包含 'demo' (如 demo-message-nest.engigu.cn)
 * 4. URL Query / Hash 参数: ?demo=1 / ?demo=true / ?mode=demo
 */

let runtimeDemoMode: boolean | null = null;

export const setRuntimeDemoMode = (isDemo: boolean) => {
  runtimeDemoMode = isDemo;
};

export const checkIsDemoMode = (): boolean => {
  if (runtimeDemoMode !== null) {
    return runtimeDemoMode;
  }

  try {
    // 1. 构建期/环境模式
    // @ts-ignore
    if (import.meta.env?.VITE_RUN_MODE === 'demo') {
      return true;
    }

    if (typeof window !== 'undefined') {
      // 2. 域名特征匹配 (例如 demo-message-nest.engigu.cn、demo.example.com 等)
      const hostname = (window.location.hostname || '').toLowerCase();
      if (hostname.includes('demo')) {
        return true;
      }

      // 3. URL search 参数匹配
      const searchParams = new URLSearchParams(window.location.search || '');
      const demoParam = searchParams.get('demo') || searchParams.get('mode');
      if (demoParam === '1' || demoParam === 'true' || demoParam === 'demo') {
        return true;
      }

      // 4. URL hash 参数匹配 (适配 Hash 路由，例如 /#/?demo=1 或 /#/login?demo=1)
      const hash = window.location.hash || '';
      if (hash.includes('?')) {
        const hashParams = new URLSearchParams(hash.substring(hash.indexOf('?') + 1));
        const hashDemoParam = hashParams.get('demo') || hashParams.get('mode');
        if (hashDemoParam === '1' || hashDemoParam === 'true' || hashDemoParam === 'demo') {
          return true;
        }
      }
    }
  } catch (e) {
    console.warn('检查演示模式异常:', e);
  }

  return false;
};
