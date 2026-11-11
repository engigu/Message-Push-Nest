export interface ChannelOption {
  id: string
  name: string
  type: string
}

export function createChannelSearchController(options: {
  search: (query: string) => Promise<ChannelOption[]>
  onOptions: (options: ChannelOption[]) => void
  onLoading: (loading: boolean) => void
  delay?: number
}): {
  search: (query: string) => void
  startComposition: () => void
  endComposition: (query: string) => void
  clear: () => void
}

export function getSelectedWayPayload(selectedChannel: ChannelOption | null): {
  way_id: string
  way_type: string
  way_name: string
} | null

export function isSelectedWayValid(
  selectedChannel: ChannelOption | null,
  selectedWayId: string | null,
): boolean
