import { ref } from 'vue'
import { toast } from 'vue-sonner'

/**
 * API 代码查看器公共逻辑 Composable
 */
export function useApiCodeViewer() {
  // 当前选中的标签
  const activeTab = ref('curl')

  // 代码语言选项
  const codeLanguages = [
    { value: 'curl', label: 'cURL', icon: '🌐' },
    { value: 'javascript', label: 'JS', icon: '🟨' },
    { value: 'python', label: 'Python', icon: '🐍' },
    { value: 'php', label: 'PHP', icon: '🐘' },
    { value: 'golang', label: 'Go', icon: '🐹' },
    { value: 'java', label: 'Java', icon: '☕' },
    { value: 'rust', label: 'Rust', icon: '🦀' }
  ]

  // 复制代码到剪贴板
  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('复制成功')
    } catch (err) {
      toast.error('复制失败')
    }
  }

  return {
    activeTab,
    codeLanguages,
    copyToClipboard
  }
}
