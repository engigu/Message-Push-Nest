<script setup lang="ts">
import type { ComboboxRootEmits, ComboboxRootProps } from "reka-ui"
import { ComboboxRoot, useForwardPropsEmits } from "reka-ui"

const props = defineProps<ComboboxRootProps>()
const emits = defineEmits<ComboboxRootEmits>()

const forwarded = useForwardPropsEmits(props, emits) as any
</script>

<template>
  <ComboboxRoot
    data-slot="combobox"
    v-bind="forwarded"
  >
    <slot />
  </ComboboxRoot>
</template>
