import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import AlertBanner from '@/components/AlertBanner.vue'

describe('AlertBanner', () => {
  it('renders nothing when alerts array is empty', () => {
    const wrapper = mount(AlertBanner, {
      props: { alerts: [], versionLabel: 'Version A: 1.0' },
    })
    expect(wrapper.find('.alert-banner').exists()).toBe(false)
  })

  it('renders alert message when one alert is provided', () => {
    const wrapper = mount(AlertBanner, {
      props: {
        alerts: [{ code: 'HIGH_VSF', message: 'VSF rate is 6.00% (threshold: 5%)' }],
        versionLabel: 'Version A: 1.0',
      },
    })
    expect(wrapper.find('.alert-banner').exists()).toBe(true)
    expect(wrapper.text()).toContain('VSF rate is 6.00% (threshold: 5%)')
    expect(wrapper.text()).toContain('HIGH_VSF')
  })

  it('renders version label in header', () => {
    const wrapper = mount(AlertBanner, {
      props: {
        alerts: [{ code: 'LOW_SAMPLE', message: 'Only 50 sessions' }],
        versionLabel: 'Version B: 2.0',
      },
    })
    expect(wrapper.text()).toContain('Version B: 2.0')
  })

  it('renders multiple alerts', () => {
    const wrapper = mount(AlertBanner, {
      props: {
        alerts: [
          { code: 'LOW_SAMPLE', message: 'Only 50 sessions' },
          { code: 'HIGH_VSF', message: 'VSF rate is 6.00%' },
        ],
        versionLabel: 'Version A: 1.0',
      },
    })
    expect(wrapper.findAll('.alert-banner__item')).toHaveLength(2)
  })
})

