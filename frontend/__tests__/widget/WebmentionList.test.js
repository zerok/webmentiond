import WebmentionList from '../../components/widget/WebmentionList.vue';
import { mount, createLocalVue } from "@vue/test-utils"
import Vuex from 'vuex';

describe('WebmentionList', () => {
  it('should render just a message if empty', () => {
    const localVue = createLocalVue();
    localVue.use(Vuex);
    const store = new Vuex.Store({
        state: {
            mentions: []
        }
    });
    const wrapper = mount(WebmentionList, {
        store,
        localVue
    });
    expect(wrapper.find('.webmention-list > ul').exists()).toBeFalsy();
    expect(wrapper.find('.webmention-list > p').exists()).toBeTruthy();
  });
});
