
export interface Theme {
    name: string;
    key: string;
    light: string;
    dark: string;
    oklch_light: string;
    oklch_dark: string;
}

export const THEMES: Theme[] = [
    {
        name: '雅致蓝',
        key: 'blue',
        light: '#3b82f6',
        dark: '#60a5fa',
        oklch_light: 'oklch(0.55 0.18 260)',
        oklch_dark: 'oklch(0.72 0.16 260)'
    },
    {
        name: '沉静翠',
        key: 'green',
        light: '#059669',
        dark: '#34d399',
        oklch_light: 'oklch(0.58 0.14 160)',
        oklch_dark: 'oklch(0.75 0.12 160)'
    },
    {
        name: '极简紫',
        key: 'purple',
        light: '#7c3aed',
        dark: '#a78bfa',
        oklch_light: 'oklch(0.52 0.22 295)',
        oklch_dark: 'oklch(0.70 0.18 295)'
    },
    {
        name: '陶韵橙',
        key: 'orange',
        light: '#ea580c',
        dark: '#fb923c',
        oklch_light: 'oklch(0.58 0.18 45)',
        oklch_dark: 'oklch(0.74 0.16 45)'
    },
    {
        name: '霜灰蓝',
        key: 'slate',
        light: '#475569',
        dark: '#94a3b8',
        oklch_light: 'oklch(0.45 0.05 255)',
        oklch_dark: 'oklch(0.70 0.04 255)'
    },
    {
        name: '极简青',
        key: 'teal',
        light: '#0d9488',
        dark: '#2dd4bf',
        oklch_light: 'oklch(0.55 0.12 190)',
        oklch_dark: 'oklch(0.75 0.10 190)'
    },
    {
        name: '胭脂红',
        key: 'rose',
        light: '#e11d48',
        dark: '#fb7185',
        oklch_light: 'oklch(0.50 0.20 15)',
        oklch_dark: 'oklch(0.70 0.18 15)'
    }
]

export const applyTheme = (themeKey: string) => {
    const theme = THEMES.find(t => t.key === themeKey) || THEMES[0]
    const root = document.documentElement

    // 注入 OKLCH 值到 CSS变量
    root.style.setProperty('--brand-light', theme.oklch_light)
    root.style.setProperty('--brand-dark', theme.oklch_dark)

    // 为了让 tailwind 生效，我们还需要更新 index.css 中定义的 --brand
    // 由于 index.css 中的 --brand 在 .dark 下有覆盖，我们需要分别处理

    // 创建或更新全局样式块来覆盖 .dark 下的变量
    let styleEl = document.getElementById('dynamic-theme-style') as HTMLStyleElement | null
    if (!styleEl) {
        styleEl = document.createElement('style')
        styleEl.id = 'dynamic-theme-style'
        document.head.appendChild(styleEl)
    }

    styleEl.innerHTML = `
    :root {
      --brand: ${theme.oklch_light};
    }
    .dark {
      --brand: ${theme.oklch_dark};
    }
  `

    localStorage.setItem('themeColor', themeKey)
}

export const getStoredTheme = (): string => {
    return localStorage.getItem('themeColor') || 'blue'
}
