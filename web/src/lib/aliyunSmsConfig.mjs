export const ALIYUN_STRUCTURED_MODE = 'structured'
export const ALIYUN_LEGACY_MODE = 'content'

export function parseTemplatePlaceholders(raw) {
  try {
    const value = typeof raw === 'string' ? JSON.parse(raw || '[]') : raw
    if (!Array.isArray(value)) return []
    const seen = new Set()
    return value
      .filter((item) => item && typeof item.key === 'string' && item.key.trim())
      .map((item) => ({
        key: item.key.trim(),
        label: typeof item.label === 'string' && item.label.trim() ? item.label.trim() : item.key.trim(),
        default: item.default == null ? '' : String(item.default),
      }))
      .filter((item) => {
        if (seen.has(item.key)) return false
        seen.add(item.key)
        return true
      })
  } catch {
    return []
  }
}

export function areTemplatePlaceholdersValid(raw) {
  if (raw == null || raw === '') return true
  try {
    return Array.isArray(typeof raw === 'string' ? JSON.parse(raw) : raw)
  } catch {
    return false
  }
}

export function createAliyunStructuredConfig(placeholders, base = {}) {
  return {
    ...base,
    template_param_mode: ALIYUN_STRUCTURED_MODE,
    selected_params: placeholders.map((item) => item.key),
    param_name_overrides: {},
  }
}

export function buildInstanceFormForChannel(channelType, placeholders, messageFormat = 'text') {
  if (channelType === 'AliyunSMS') {
    return createAliyunStructuredConfig(placeholders, { allowMultiRecip: false, templ_type: messageFormat })
  }
  return { allowMultiRecip: false, templ_type: messageFormat }
}

export function normalizeAliyunConfig(config = {}) {
  const mode = config.template_param_mode || ALIYUN_LEGACY_MODE
  return {
    ...config,
    template_param_mode: mode,
    selected_params: Array.isArray(config.selected_params) ? [...config.selected_params] : [],
    param_name_overrides: config.param_name_overrides && typeof config.param_name_overrides === 'object'
      ? { ...config.param_name_overrides }
      : {},
  }
}

export function toggleAliyunParam(config, key, checked) {
  const selected = new Set(config.selected_params || [])
  if (checked) selected.add(key)
  else selected.delete(key)
  const overrides = { ...(config.param_name_overrides || {}) }
  if (!checked) delete overrides[key]
  return { ...config, selected_params: [...selected], param_name_overrides: overrides }
}

export function setAliyunParamOverride(config, key, value) {
  const overrides = { ...(config.param_name_overrides || {}) }
  const normalized = value.trim()
  if (!normalized || normalized === key) delete overrides[key]
  else overrides[key] = normalized
  return { ...config, param_name_overrides: overrides }
}

export function getAliyunPlaceholderDiff(config, placeholders) {
  const current = new Set(placeholders.map((item) => item.key))
  const selected = new Set(config.selected_params || [])
  return {
    added: placeholders.filter((item) => !selected.has(item.key)),
    removed: [...selected].filter((key) => !current.has(key)),
  }
}

export function validateAliyunConfig(config, placeholders, placeholdersValid = true) {
  if (!String(config.template_code || '').trim()) return '请填写阿里云短信 TemplateCode'
  if (config.template_param_mode !== ALIYUN_STRUCTURED_MODE) return ''

  if (!placeholdersValid) return '当前消息模板的 Placeholders 格式无效，请先修复模板变量定义'

  if (!Array.isArray(config.selected_params)) return '参数集合无效，请重新选择短信变量'
  const placeholderKeys = new Set(placeholders.map((item) => item.key))
  const selected = config.selected_params
  if (new Set(selected).size !== selected.length) return '已选业务变量存在重复项'
  const invalid = selected.find((key) => !placeholderKeys.has(key))
  if (invalid) return `变量“${invalid}”已从当前模板删除或改名，请移除后再保存`

  const targets = new Map()
  for (const key of selected) {
    const override = config.param_name_overrides?.[key]
    if (override !== undefined && (!String(override).trim() || override.trim() !== override)) {
      return `变量“${key}”的阿里云参数名不能为空或包含首尾空格`
    }
    if (override && override === key) return `变量“${key}”与阿里云参数同名，请清除覆盖映射`
    const target = override || key
    if (targets.has(target)) return `变量“${key}”与“${targets.get(target)}”映射到相同的阿里云参数名“${target}”`
    targets.set(target, key)
  }
  const unusedOverride = Object.keys(config.param_name_overrides || {}).find((key) => !selected.includes(key))
  if (unusedOverride) return `变量“${unusedOverride}”未勾选发送，请移除它的参数映射`
  return ''
}

const editableConfigFields = [
  'phone_number', 'template_code', 'template_param_mode', 'selected_params',
  'param_name_overrides', 'allowMultiRecip',
]

export function toAliyunEditableConfig(config) {
  return Object.fromEntries(editableConfigFields
    .filter((key) => Object.prototype.hasOwnProperty.call(config, key))
    .map((key) => [key, config[key]]))
}

export function buildAliyunTestPayload({ scenario, instanceId, templateId, wayId, config, testPhone, placeholders }) {
  const payload = {
    test_phone: testPhone,
    placeholders: { ...placeholders },
  }
  if (scenario === 'saved') {
    payload.instance_id = instanceId
    return payload
  }
  if (scenario === 'edit') payload.instance_id = instanceId
  payload.template_id = templateId
  payload.way_id = wayId
  payload.config = toAliyunEditableConfig(config)
  return payload
}

export function getAliyunTestValues(config, placeholders, initial = {}) {
  const selected = config.template_param_mode === ALIYUN_STRUCTURED_MODE
    ? new Set(config.selected_params || [])
    : new Set(placeholders.map((item) => item.key))
  return Object.fromEntries(placeholders
    .filter((item) => selected.has(item.key))
    .map((item) => [item.key, initial[item.key] ?? item.default ?? '']))
}

export function getAliyunConfigSummary(rawConfig) {
  try {
    const config = typeof rawConfig === 'string' ? JSON.parse(rawConfig || '{}') : rawConfig || {}
    const overrides = config.param_name_overrides && typeof config.param_name_overrides === 'object'
      ? config.param_name_overrides
      : {}
    return {
      templateCode: config.template_code || '',
      mode: config.template_param_mode || ALIYUN_LEGACY_MODE,
      selectedCount: config.template_param_mode === ALIYUN_STRUCTURED_MODE
        ? (Array.isArray(config.selected_params) ? config.selected_params.length : 0)
        : null,
      overrideCount: config.template_param_mode === ALIYUN_STRUCTURED_MODE ? Object.keys(overrides).length : null,
    }
  } catch {
    return { templateCode: '', mode: ALIYUN_LEGACY_MODE, selectedCount: null, overrideCount: null }
  }
}
