import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import KPICard from '@/components/KPICard.vue'

describe('KPICard', () => {
  it('renders version A and B values', () => {
    const wrapper = mount(KPICard, {
      props: { name: 'VSF Rate', valueA: 0.012, valueB: 0.034, winner: 'A' },
    })
    expect(wrapper.text()).toContain('0.012')
    expect(wrapper.text()).toContain('0.034')
    expect(wrapper.text()).toContain('VSF Rate')
  })

  it('applies winner class to A and loser class to B when A wins', () => {
    const wrapper = mount(KPICard, {
      props: { name: 'VSF Rate', valueA: 0.01, valueB: 0.05, winner: 'A' },
    })
    const values = wrapper.findAll('.kpi-card__value')
    expect(values[0].classes()).toContain('kpi-card__value--winner')
    expect(values[1].classes()).toContain('kpi-card__value--loser')
  })

  it('applies winner class to B and loser class to A when B wins', () => {
    const wrapper = mount(KPICard, {
      props: { name: 'VSF Rate', valueA: 0.05, valueB: 0.01, winner: 'B' },
    })
    const values = wrapper.findAll('.kpi-card__value')
    expect(values[0].classes()).toContain('kpi-card__value--loser')
    expect(values[1].classes()).toContain('kpi-card__value--winner')
  })

  it('applies neither winner nor loser class on tie', () => {
    const wrapper = mount(KPICard, {
      props: { name: 'Play Rate', valueA: 0.9, valueB: 0.9, winner: 'tie' },
    })
    const values = wrapper.findAll('.kpi-card__value')
    expect(values[0].classes()).not.toContain('kpi-card__value--winner')
    expect(values[0].classes()).not.toContain('kpi-card__value--loser')
    expect(values[1].classes()).not.toContain('kpi-card__value--winner')
    expect(values[1].classes()).not.toContain('kpi-card__value--loser')
  })

  it('renders optional unit suffix', () => {
    const wrapper = mount(KPICard, {
      props: { name: 'Avg VST', valueA: 1.5, valueB: 2.0, winner: 'A', unit: 's' },
    })
    expect(wrapper.text()).toContain('1.500s')
  })
})

