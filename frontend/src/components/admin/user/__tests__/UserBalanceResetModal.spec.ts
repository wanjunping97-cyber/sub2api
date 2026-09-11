import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import type { AdminUser } from '@/types'

const apiMocks = vi.hoisted(() => ({
  resetBalance: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      resetBalance: apiMocks.resetBalance
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('@/components/common/BaseDialog.vue', () => ({
  default: {
    name: 'BaseDialog',
    props: ['show', 'title', 'width'],
    template: '<div v-if="show"><slot /><slot name="footer" /></div>'
  }
}))

import UserBalanceResetModal from '../UserBalanceResetModal.vue'

const createUser = (overrides: Partial<AdminUser> = {}): AdminUser => ({
  id: 7,
  username: 'reset-user',
  email: 'reset@example.com',
  role: 'user',
  balance: 12,
  concurrency: 1,
  status: 'active',
  allowed_groups: [],
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-09-11T00:00:00Z',
  updated_at: '2026-09-11T00:00:00Z',
  notes: '',
  ...overrides
})

async function mountAndOpen(user: AdminUser) {
  const wrapper = mount(UserBalanceResetModal, {
    props: { show: false, user }
  })
  await wrapper.setProps({ show: true })
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  apiMocks.resetBalance.mockResolvedValue(createUser({ balance: 80, last_balance_reset_value: 80 }))
})

describe('UserBalanceResetModal', () => {
  it('prefills the last reset face value when present', async () => {
    const wrapper = await mountAndOpen(createUser({
      balance: 3,
      last_balance_reset_value: 80
    }))

    const input = wrapper.get('[data-test="reset-balance-amount"]').element as HTMLInputElement
    expect(input.value).toBe('80')
    expect(wrapper.text()).toContain('admin.users.resetBalanceDefaultHint')
  })

  it('falls back to the current balance when there is no last reset value', async () => {
    const wrapper = await mountAndOpen(createUser({ balance: 12 }))

    const input = wrapper.get('[data-test="reset-balance-amount"]').element as HTMLInputElement
    expect(input.value).toBe('12')
    expect(wrapper.text()).not.toContain('admin.users.resetBalanceDefaultHint')
  })

  it('calls resetBalance with the overlay amount and emits success', async () => {
    const wrapper = await mountAndOpen(createUser({
      last_balance_reset_value: 80
    }))

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(apiMocks.resetBalance).toHaveBeenCalledWith(7, 80, '')
    expect(wrapper.emitted('success')).toBeTruthy()
    expect(wrapper.emitted('close')).toBeTruthy()
  })
})
