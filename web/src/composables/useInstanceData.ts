import { ref, computed, watch } from 'vue'
import { request } from '@/api/api'

/**
 * 实例数据管理 Composable
 * 用于任务和模板的实例数据加载和处理
 */
export function useInstanceData(
  type: 'task' | 'template',
  dataRef: any,
  openRef: any
) {
  // 实例数据
  const instances = ref<any[]>([])

  // 检查是否有支持动态接收者的实例
  const hasDynamicRecipientInstance = computed(() => {
    if (instances.value && Array.isArray(instances.value)) {
      return instances.value.some((ins: any) => {
        try {
          const config = typeof ins.config === 'string' ? JSON.parse(ins.config) : ins.config
          return config?.allowMultiRecip === true
        } catch {
          return false
        }
      })
    }
    return false
  })

  // 获取已启用的实例渠道名称列表
  const enabledChannelNames = computed(() => {
    if (instances.value && Array.isArray(instances.value)) {
      return instances.value
        .filter((ins: any) => ins.enable === 1)
        .map((ins: any) => ins.way_name)
    }
    return []
  })

  // 加载实例数据
  const loadInstances = async () => {
    const id = dataRef.value?.id
    if (!id) return

    try {
      let response
      if (type === 'template') {
        response = await request.get('/templates/ins/get', {
          params: { id }
        })
        instances.value = response.data.data.ins_list || []
      } else {
        response = await request.get('/sendtasks/ins/gettask', {
          params: { id }
        })
        instances.value = response.data.data.ins_data || []
      }
    } catch (err) {
      console.error(`加载${type === 'template' ? '模板' : '任务'}实例失败:`, err)
      instances.value = []
    }
  }

  // 监听弹窗打开状态
  watch(
    () => openRef.value,
    async (newVal: boolean) => {
      if (newVal) {
        await loadInstances()
      } else {
        instances.value = []
      }
    }
  )

  return {
    instances,
    hasDynamicRecipientInstance,
    enabledChannelNames,
    loadInstances
  }
}
