import Index from '../../components/Index.vue';
import Vuex from 'vuex';
import { mount } from "@vue/test-utils"

describe('Index', () => {
  it('should render no list if empty', () => {
    const store = new Vuex.Store({
        state: {
            mentions: []
        },
        actions: {
          getMentions: () => {},
        }
    });
    const wrapper = mount(Index, {
        global: {
          plugins: [store]
        }
    });
    expect(wrapper.find('.mention-list').exists()).toBeFalsy();
  });

  it('should render a delete button for every mention', () => {
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
        global: {
          plugins: [store]
        }
    });
    expect(wrapper.find('.mention-list').exists()).toBeTruthy();
    expect(wrapper.find('.mention-list li button.button--delete').exists()).toBeTruthy();
  });
});
