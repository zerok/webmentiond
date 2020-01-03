import Loading from '../../components/Loading.vue';
import Vue from 'vue';

describe('Loading', () => {
  it('should result in a div with the class "loading"', () => {
    const vm = new Vue(Loading).$mount();
    expect(vm.$el.getAttribute('class')).toBe('loading');
  });
});
