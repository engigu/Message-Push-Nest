import assert from 'node:assert/strict'
import test from 'node:test'
import {
  createChannelSearchController,
  getSelectedWayPayload,
  isSelectedWayValid,
} from './channelSearchController.mjs'

const wait = (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds))

function createHarness(search, delay = 5) {
  const state = { options: [], loading: false }
  const controller = createChannelSearchController({
    search,
    delay,
    onOptions: (options) => { state.options = options },
    onLoading: (loading) => { state.loading = loading },
  })
  return { controller, state }
}

test('searches Chinese and English queries and publishes remote results', async () => {
  const queries = []
  const result = [{ id: 'way-1', name: '阿里云短信', type: 'aliyun_sms' }]
  const { controller, state } = createHarness(async (query) => {
    queries.push(query)
    return result
  })

  controller.search('阿里云')
  await wait(12)
  assert.deepEqual(state.options, result)
  controller.search('aliyun')
  await wait(12)
  assert.deepEqual(queries, ['阿里云', 'aliyun'])
  assert.deepEqual(state.options, result)
  assert.equal(state.loading, false)
})

test('debounces rapid typing and backspace to the latest query', async () => {
  const queries = []
  const { controller } = createHarness(async (query) => {
    queries.push(query)
    return []
  }, 12)

  controller.search('aliy')
  await wait(3)
  controller.search('aliyu')
  await wait(3)
  controller.search('aliyun')
  await wait(18)
  controller.search('aliyu')
  await wait(18)

  assert.deepEqual(queries, ['aliyun', 'aliyu'])
})

test('clearing a query clears results and loading state immediately', async () => {
  const result = [{ id: 'way-1', name: '阿里云短信', type: 'aliyun_sms' }]
  const { controller, state } = createHarness(async () => result)
  controller.search('阿里云')
  await wait(12)
  assert.deepEqual(state.options, result)

  controller.search('')
  assert.deepEqual(state.options, [])
  assert.equal(state.loading, false)
})

test('ignores older responses that arrive after the newest query', async () => {
  const pending = new Map()
  const { controller, state } = createHarness((query) => new Promise((resolve) => {
    pending.set(query, resolve)
  }), 0)

  controller.search('old')
  await wait(2)
  controller.search('new')
  await wait(2)
  const newest = [{ id: 'new-id', name: 'Newest', type: 'email' }]
  pending.get('new')(newest)
  await wait(0)
  pending.get('old')([{ id: 'old-id', name: 'Old', type: 'email' }])
  await wait(0)

  assert.deepEqual(state.options, newest)
  assert.equal(state.loading, false)
})

test('does not search intermediate IME composition values and searches the final value', async () => {
  const queries = []
  const { controller } = createHarness(async (query) => {
    queries.push(query)
    return []
  })

  controller.startComposition()
  controller.search('a')
  controller.search('阿里')
  await wait(8)
  assert.deepEqual(queries, [])

  controller.endComposition('阿里云')
  await wait(12)
  assert.deepEqual(queries, ['阿里云'])
})

test('uses the selected non-first option and rejects missing selections', () => {
  const options = [
    { id: 'first', name: 'Email', type: 'email' },
    { id: 'second', name: 'Aliyun SMS', type: 'aliyun_sms' },
  ]

  assert.deepEqual(getSelectedWayPayload(options[1]), {
    way_id: 'second',
    way_type: 'aliyun_sms',
    way_name: 'Aliyun SMS',
  })
  assert.equal(getSelectedWayPayload(null), null)
  assert.equal(getSelectedWayPayload({ id: 'id-only' }), null)
  assert.equal(isSelectedWayValid(null, null), false)
  assert.equal(isSelectedWayValid(options[1], 'first'), false)
  assert.equal(isSelectedWayValid(options[1], 'second'), true)
})
