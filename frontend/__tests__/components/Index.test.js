import Index from '../../components/Index.vue';
import { mount, createLocalVue } from "@vue/test-utils"
import Vuex from 'vuex';

describe('Index', () => {
  it('should render no list if empty', () => {
    const localVue = createLocalVue();
    localVue.use(Vuex);
    const store = new Vuex.Store({
        state: {
            mentions: []
        },
        actions: {
          getMentions: () => {},
        }
    });
    const wrapper = mount(Index, {
        store,
        localVue
    });
    expect(wrapper.find('.mention-list').exists()).toBeFalsy();
  });

  it('should render a delete button for every mention', () => {
    const localVue = createLocalVue();
    localVue.use(Vuex);
    const store = new Vuex.Store({
        state: {
            mentions: [{
              status: 'rejected'
            }],
            pagingInfo: {
              total: 1,
            }
        },
        actions: {
          getMentions: () => {},
        }
    });
    const wrapper = mount(Index, {
        store,
        localVue
    });
    expect(wrapper.find('.mention-list').exists()).toBeTruthy();
    expect(wrapper.find('.mention-list li button.button--delete').exists()).toBeTruthy();
  });
});
