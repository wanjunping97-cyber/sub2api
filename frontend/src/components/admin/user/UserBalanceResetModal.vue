<template>
  <BaseDialog :show="show" :title="t('admin.users.resetBalance')" width="narrow" @close="$emit('close')">
    <form v-if="user" id="balance-reset-form" @submit.prevent="handleSubmit" class="space-y-5">
      <div class="flex items-center gap-3 rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
        <div class="flex h-10 w-10 items-center justify-center rounded-full bg-primary-100">
          <span class="text-lg font-medium text-primary-700">{{ user.email.charAt(0).toUpperCase() }}</span>
        </div>
        <div class="flex-1">
          <p class="font-medium text-gray-900 dark:text-gray-100">{{ user.email }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.users.currentBalance') }}: ${{ formatBalance(user.balance) }}
          </p>
        </div>
      </div>
      <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('admin.users.resetBalanceHint') }}</p>
      <div>
        <label class="input-label">{{ t('admin.users.resetBalanceAmount') }}</label>
        <div class="relative">
          <div class="absolute left-3 top-1/2 -translate-y-1/2 font-medium text-gray-500">$</div>
          <input
            v-model.number="form.amount"
            data-test="reset-balance-amount"
            type="number"
            step="any"
            min="0"
            required
            class="input pl-8"
          />
        </div>
        <p v-if="usedLastResetValue" class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.users.resetBalanceDefaultHint') }}
        </p>
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.notes') }}</label>
        <textarea
          v-model="form.notes"
          rows="3"
          class="input"
          :placeholder="t('admin.users.resetBalanceNotesPlaceholder')"
        ></textarea>
      </div>
      <div
        v-if="form.amount > 0"
        class="rounded-xl border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-950"
      >
        <div class="flex items-center justify-between text-sm">
          <span class="text-gray-700 dark:text-gray-300">{{ t('admin.users.newBalance') }}:</span>
          <span class="font-bold text-gray-900 dark:text-gray-100">${{ formatBalance(form.amount) }}</span>
        </div>
      </div>
    </form>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button @click="$emit('close')" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button
          type="submit"
          form="balance-reset-form"
          data-test="confirm-reset-balance"
          :disabled="submitting || !form.amount"
          class="btn bg-amber-600 text-white hover:bg-amber-700"
        >
          {{ submitting ? t('admin.users.resettingBalance') : t('admin.users.confirmResetBalance') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits<{ close: []; success: [user: AdminUser] }>()
const { t } = useI18n()
const appStore = useAppStore()

const submitting = ref(false)
const usedLastResetValue = ref(false)
const form = reactive({ amount: 0, notes: '' })

const defaultResetAmount = (user: AdminUser | null): { amount: number; fromLast: boolean } => {
  if (!user) return { amount: 0, fromLast: false }
  const last = user.last_balance_reset_value
  if (typeof last === 'number' && last > 0) {
    return { amount: last, fromLast: true }
  }
  if (user.balance > 0) {
    return { amount: user.balance, fromLast: false }
  }
  return { amount: 0, fromLast: false }
}

watch(
  () => props.show,
  (visible) => {
    if (!visible) return
    const preset = defaultResetAmount(props.user)
    form.amount = preset.amount
    form.notes = ''
    usedLastResetValue.value = preset.fromLast
  }
)

const formatBalance = (value: number) => {
  if (value === 0) return '0.00'
  const formatted = value.toFixed(8).replace(/\.?0+$/, '')
  const parts = formatted.split('.')
  if (parts.length === 1) return formatted + '.00'
  if (parts[1].length === 1) return formatted + '0'
  return formatted
}

const handleSubmit = async () => {
  if (!props.user) return
  if (!form.amount || form.amount <= 0) {
    appStore.showError(t('admin.users.amountRequired'))
    return
  }
  submitting.value = true
  try {
    const updated = await adminAPI.users.resetBalance(props.user.id, form.amount, form.notes)
    appStore.showSuccess(t('admin.users.resetBalanceSuccess'))
    emit('success', updated)
    emit('close')
  } catch (e: any) {
    console.error('Failed to reset balance:', e)
    appStore.showError(e.response?.data?.detail || t('common.error'))
  } finally {
    submitting.value = false
  }
}
</script>
