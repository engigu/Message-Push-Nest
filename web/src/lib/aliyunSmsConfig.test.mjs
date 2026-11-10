import assert from 'node:assert/strict'
import test from 'node:test'
import {
  areTemplatePlaceholdersValid,
  buildInstanceFormForChannel,
  buildAliyunTestPayload,
  createAliyunStructuredConfig,
  getAliyunConfigSummary,
  getAliyunPlaceholderDiff,
  getAliyunTestValues,
  normalizeAliyunConfig,
  parseTemplatePlaceholders,
  setAliyunParamOverride,
  toAliyunEditableConfig,
  toggleAliyunParam,
  validateAliyunConfig,
} from './aliyunSmsConfig.mjs'

const placeholders = parseTemplatePlaceholders('[{"key":"name","label":"姓名","default":"测试用户"},{"key":"code","label":"验证码","default":"123456"}]')

test('new instances select every template variable and start with same-name mappings', () => {
  const config = createAliyunStructuredConfig(placeholders, { template_code: 'SMS_TEST' })
  assert.deepEqual(config.selected_params, ['name', 'code'])
  assert.deepEqual(config.param_name_overrides, {})
  assert.equal(validateAliyunConfig(config, placeholders), '')
})

test('malformed template placeholder definitions block structured saves', () => {
  const config = createAliyunStructuredConfig(placeholders, { template_code: 'SMS_TEST' })
  assert.equal(areTemplatePlaceholdersValid('{broken'), false)
  assert.match(validateAliyunConfig(config, [], areTemplatePlaceholdersValid('{broken')), /Placeholders 格式无效/)
})

test('switching away from Aliyun clears its dedicated config while other channels keep their basic defaults', () => {
  const aliyun = buildInstanceFormForChannel('AliyunSMS', placeholders, 'text')
  assert.deepEqual(aliyun.selected_params, ['name', 'code'])
  const email = buildInstanceFormForChannel('Email', placeholders, 'html')
  assert.deepEqual(email, { allowMultiRecip: false, templ_type: 'html' })
})

test('unchecked variables are excluded and their override is removed', () => {
  const config = setAliyunParamOverride(createAliyunStructuredConfig(placeholders), 'name', 'customer')
  const updated = toggleAliyunParam(config, 'name', false)
  assert.deepEqual(updated.selected_params, ['code'])
  assert.equal('name' in updated.param_name_overrides, false)
})

test('only different names are stored and conflicting targets are rejected', () => {
  const config = createAliyunStructuredConfig(placeholders, { template_code: 'SMS_TEST' })
  const mapped = setAliyunParamOverride(config, 'name', 'customer')
  assert.deepEqual(mapped.param_name_overrides, { name: 'customer' })
  assert.equal(validateAliyunConfig(mapped, placeholders), '')
  const conflicting = setAliyunParamOverride(mapped, 'code', 'customer')
  assert.match(validateAliyunConfig(conflicting, placeholders), /相同的阿里云参数名/)
  assert.deepEqual(setAliyunParamOverride(mapped, 'name', 'name').param_name_overrides, {})
})

test('editing keeps the selected set fixed when template adds a variable', () => {
  const config = { template_code: 'SMS', template_param_mode: 'structured', selected_params: ['name'], param_name_overrides: {} }
  const changed = [...placeholders, { key: 'count', label: '数量', default: '' }]
  assert.deepEqual(getAliyunPlaceholderDiff(config, changed).added.map((item) => item.key), ['code', 'count'])
  assert.deepEqual(config.selected_params, ['name'])
})

test('deleted or renamed selected variables are reported and rejected until removed', () => {
  const config = { template_code: 'SMS', template_param_mode: 'structured', selected_params: ['old_name'], param_name_overrides: { old_name: 'legacy_name' } }
  assert.deepEqual(getAliyunPlaceholderDiff(config, placeholders).removed, ['old_name'])
  assert.match(validateAliyunConfig(config, placeholders), /已从当前模板删除或改名/)
  const cleaned = toggleAliyunParam(config, 'old_name', false)
  assert.equal(validateAliyunConfig(cleaned, placeholders), '')
})

test('legacy content config stays in compatibility mode and does not gain structured parameters', () => {
  const legacy = normalizeAliyunConfig({ template_code: 'SMS_OLD', phone_number: '15500000000' })
  assert.equal(legacy.template_param_mode, 'content')
  assert.deepEqual(legacy.selected_params, [])
  assert.deepEqual(getAliyunTestValues(legacy, placeholders), { name: '测试用户', code: '123456' })
  assert.deepEqual(toAliyunEditableConfig(legacy), {
    phone_number: '15500000000', template_code: 'SMS_OLD', template_param_mode: 'content',
    selected_params: [], param_name_overrides: {},
  })
})

test('new and edit draft tests send the current config; saved tests only send instance id', () => {
  const config = {
    phone_number: '15500000000', template_code: 'SMS_DRAFT', template_param_mode: 'structured',
    selected_params: ['name'], param_name_overrides: { name: 'customer' }, access_key_secret: 'never-send',
  }
  const common = { instanceId: 'IN1', templateId: 'TP1', wayId: 'WY1', config, testPhone: '15555550123', placeholders: { name: 'draft' } }
  const created = buildAliyunTestPayload({ ...common, scenario: 'new' })
  assert.deepEqual(created, {
    test_phone: '15555550123', placeholders: { name: 'draft' }, template_id: 'TP1', way_id: 'WY1',
    config: { phone_number: '15500000000', template_code: 'SMS_DRAFT', template_param_mode: 'structured', selected_params: ['name'], param_name_overrides: { name: 'customer' } },
  })
  const edited = buildAliyunTestPayload({ ...common, scenario: 'edit' })
  assert.equal(edited.instance_id, 'IN1')
  assert.equal(edited.config.template_code, 'SMS_DRAFT')
  const saved = buildAliyunTestPayload({ ...common, scenario: 'saved' })
  assert.deepEqual(saved, { test_phone: '15555550123', placeholders: { name: 'draft' }, instance_id: 'IN1' })
  assert.equal('config' in saved, false)
  assert.equal(JSON.stringify(created).includes('never-send'), false)
})

test('response summary shows mode and selected/override counts without serializing credentials', () => {
  const summary = getAliyunConfigSummary(JSON.stringify({ template_code: 'SMS', template_param_mode: 'structured', selected_params: ['name'], param_name_overrides: { name: 'customer' }, access_key_secret: 'hidden' }))
  assert.deepEqual(summary, { templateCode: 'SMS', mode: 'structured', selectedCount: 1, overrideCount: 1 })
})
