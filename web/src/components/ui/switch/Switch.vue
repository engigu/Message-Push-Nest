<script setup lang="ts">
import type { SwitchRootEmits, SwitchRootProps } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { computed } from "vue"
import {
  SwitchRoot,
  SwitchThumb,
  useForwardPropsEmits,
} from "reka-ui"
import { cn } from "@/lib/utils"

interface ExtendedSwitchProps extends SwitchRootProps {
  class?: HTMLAttributes["class"]
  checked?: boolean
}

const props = defineProps<ExtendedSwitchProps>()

const emits = defineEmits<SwitchRootEmits & {
  'update:checked': [value: boolean]
}>()

const delegatedProps = computed(() => {
  const { class: _, checked, ...rest } = props
  return {
    ...rest,
    modelValue: rest.modelValue !== undefined ? rest.modelValue : checked,
  }
})

const forwarded = useForwardPropsEmits(delegatedProps, emits)

const handleUpdateModelValue = (val: boolean) => {
  emits('update:checked', val)
}
</script>

<template>
  <SwitchRoot
    data-slot="switch"
    v-bind="forwarded"
    @update:model-value="handleUpdateModelValue"
    :class="cn(
      'peer data-[state=checked]:bg-primary data-[state=unchecked]:bg-input focus-visible:border-ring focus-visible:ring-ring/50 dark:data-[state=unchecked]:bg-input/80 inline-flex h-[1.15rem] w-8 shrink-0 items-center rounded-full border border-transparent shadow-xs transition-all outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50',
      props.class,
    )"
  >
    <SwitchThumb
      data-slot="switch-thumb"
      :class="cn('bg-background dark:data-[state=unchecked]:bg-foreground dark:data-[state=checked]:bg-primary-foreground pointer-events-none block size-4 rounded-full ring-0 transition-transform data-[state=checked]:translate-x-[calc(100%-2px)] data-[state=unchecked]:translate-x-0')"
    >
      <slot name="thumb" />
    </SwitchThumb>
  </SwitchRoot>
</template>
