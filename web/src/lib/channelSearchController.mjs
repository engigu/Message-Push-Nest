export function createChannelSearchController({
  search,
  onOptions,
  onLoading,
  delay = 300,
}) {
  let timer
  let requestVersion = 0
  let composing = false

  const invalidate = () => {
    requestVersion += 1
    if (timer !== undefined) {
      clearTimeout(timer)
      timer = undefined
    }
    return requestVersion
  }

  const clear = () => {
    invalidate()
    onOptions([])
    onLoading(false)
  }

  const runSearch = (query) => {
    const normalizedQuery = query.trim()
    const version = invalidate()

    if (!normalizedQuery || composing) {
      if (!normalizedQuery) {
        onOptions([])
        onLoading(false)
      }
      return
    }

    onLoading(true)
    timer = setTimeout(async () => {
      timer = undefined
      try {
        const options = await search(normalizedQuery)
        if (version === requestVersion && !composing) onOptions(options)
      } catch (error) {
        if (version === requestVersion && !composing) {
          onOptions([])
          console.error('搜索渠道失败', error)
        }
      } finally {
        if (version === requestVersion && !composing) onLoading(false)
      }
    }, delay)
  }

  return {
    search(query) {
      runSearch(query)
    },
    startComposition() {
      composing = true
      clear()
    },
    endComposition(query) {
      composing = false
      runSearch(query)
    },
    clear,
  }
}

export function getSelectedWayPayload(selectedChannel) {
  if (!selectedChannel?.id || !selectedChannel?.type || !selectedChannel?.name) return null
  return {
    way_id: selectedChannel.id,
    way_type: selectedChannel.type,
    way_name: selectedChannel.name,
  }
}

export function isSelectedWayValid(selectedChannel, selectedWayId) {
  return Boolean(getSelectedWayPayload(selectedChannel) && selectedChannel.id === selectedWayId)
}
