import {createApp} from 'vue';
import Widget from './components/widget/Widget.vue';
import Vuex from 'vuex';

const store = new Vuex.Store({
    state: {
        mentions: null
    },
    mutations: {
        setMentions(state, mentions) {
            state.mentions = mentions
        }
    },
    actions: {
        async fetchMentions({commit}, {endpoint, target}) {
            const resp = await fetch(`${endpoint}/get?target=${target}`);
            const data = await resp.json();
            commit('setMentions', data);
        }
    }
});

const setup = () => {
    const container = document.querySelector('.webmentions-container');
    if (typeof container === 'undefined') {
        return;
    }
    const mentions = container.innerText !== '' ? JSON.parse(container.innerText) : null;
    const title = container.dataset.title || 'Mentions';
    const endpoint = container.dataset.endpoint;
    const target = container.dataset.target;
    const showRSVPSummary = container.dataset.showRSVPSummary === 'yes';

    const app = createApp(Widget, {
        mentions,
        endpoint,
        title,
        showRSVPSummary,
        target
    });
    
    app.use(store);
    app.mount(container);
};

setup();
