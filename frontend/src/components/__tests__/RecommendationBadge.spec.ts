import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import RecommendationBadge from '@/components/RecommendationBadge.vue'
import type { Recommendation } from '@/services/api'

function makeRec(overrides: Partial<Recommendation> = {}): Recommendation {
  return {
    winner: '1.0',
    wins_a: 4,
    wins_b: 2,
    reason: 'Version 1.0 wins on 4/6 metrics',
    dimensions: [],
    ...overrides,
  }
}

describe('RecommendationBadge', () => {
  it('displays winner version label', () => {
    const wrapper = mount(RecommendationBadge, {
      props: { recommendation: makeRec({ winner: '7.1.13' }) },
    })
    expect(wrapper.text()).toContain('7.1.13')
    expect(wrapper.classes()).toContain('rec-badge--winner')
  })

  it('displays reason text', () => {
    const wrapper = mount(RecommendationBadge, {
      props: { recommendation: makeRec({ reason: 'Version 1.0 wins on 4/6 metrics' }) },
    })
    expect(wrapper.text()).toContain('Version 1.0 wins on 4/6 metrics')
  })

  it('displays tie message when winner is empty', () => {
    const wrapper = mount(RecommendationBadge, {
      props: { recommendation: makeRec({ winner: '', wins_a: 3, wins_b: 3, reason: 'Tie (3/6 each)' }) },
    })
    expect(wrapper.text()).toContain('Tie')
    expect(wrapper.classes()).toContain('rec-badge--tie')
    expect(wrapper.classes()).not.toContain('rec-badge--winner')
  })

  it('displays wins breakdown', () => {
    const wrapper = mount(RecommendationBadge, {
      props: { recommendation: makeRec({ wins_a: 4, wins_b: 2 }) },
    })
    expect(wrapper.text()).toContain('A: 4 wins')
    expect(wrapper.text()).toContain('B: 2 wins')
  })
})

