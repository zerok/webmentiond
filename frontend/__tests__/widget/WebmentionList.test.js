import WebmentionList from '../../components/widget/WebmentionList.vue';
import { mount } from "@vue/test-utils"
import Vuex from 'vuex';

describe('WebmentionList', () => {
  it('should render just a message if empty', () => {
    const store = new Vuex.Store({
        state: {
            mentions: []
        }
    });
    const wrapper = mount(WebmentionList, {
        global: {
            plugins: [store]
        }
    });
    expect(wrapper.find('.webmention-list > ul').exists()).toBeFalsy();
    expect(wrapper.find('.webmention-list > p').exists()).toBeTruthy();
  });
  it('should render a heart for a like', () => {
    const store = new Vuex.Store({
        state: {
            mentions: [{
                type: 'like',
                source: 'https://source.com',
                target: 'https://target.com',
                title: 'A mention'
            }]
        }
    });
    const wrapper = mount(WebmentionList, {
        global: {
            plugins: [store]
        }
    });
    expect(wrapper.find('.webmention-list > ul').exists()).toBeTruthy();
    expect(wrapper.find('.webmention-list > ul > li i.fa-heart').exists()).toBeTruthy();
  });
});
