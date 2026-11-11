export const ALIYUN_STRUCTURED_MODE: 'structured'
export const ALIYUN_LEGACY_MODE: 'content'

export interface TemplatePlaceholder {
  key: string
  label: string
  default: string
}

export interface AliyunConfig {
  phone_number?: string
  template_code?: string
  template_param_mode?: string
  selected_params?: string[]
  param_name_overrides?: Record<string, string>
  allowMultiRecip?: boolean
  [key: string]: any
}

export function areTemplatePlaceholdersValid(raw: string | TemplatePlaceholder[] | undefined): boolean
export function parseTemplatePlaceholders(raw: string | TemplatePlaceholder[] | undefined): TemplatePlaceholder[]
export function createAliyunStructuredConfig(placeholders: TemplatePlaceholder[], base?: AliyunConfig): AliyunConfig
export function buildInstanceFormForChannel(channelType: string, placeholders: TemplatePlaceholder[], messageFormat?: string): AliyunConfig
export function normalizeAliyunConfig(config?: AliyunConfig): AliyunConfig
export function toggleAliyunParam(config: AliyunConfig, key: string, checked: boolean): AliyunConfig
export function setAliyunParamOverride(config: AliyunConfig, key: string, value: string): AliyunConfig
export function getAliyunPlaceholderDiff(config: AliyunConfig, placeholders: TemplatePlaceholder[]): { added: TemplatePlaceholder[]; removed: string[] }
export function validateAliyunConfig(config: AliyunConfig, placeholders: TemplatePlaceholder[], placeholdersValid?: boolean): string
export function toAliyunEditableConfig(config: AliyunConfig): AliyunConfig
export function buildAliyunTestPayload(args: {
  scenario: 'new' | 'edit' | 'saved'
  instanceId: string
  templateId: string
  wayId: string
  config: AliyunConfig
  testPhone: string
  placeholders: Record<string, any>
}): Record<string, any>
export function getAliyunTestValues(config: AliyunConfig, placeholders: TemplatePlaceholder[], initial?: Record<string, any>): Record<string, string>
export function getAliyunConfigSummary(rawConfig: string | AliyunConfig): {
  templateCode: string
  mode: string
  selectedCount: number | null
  overrideCount: number | null
}
