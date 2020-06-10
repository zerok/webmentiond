import Vue from 'vue';
import Widget from './components/widget/Widget.vue';
import Vuex from 'vuex';

Vue.use(Vuex);

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

var app = new Vue({
    el: '.webmentions-container',
    store: store,
    components: {
        Widget
    },
    beforeMount: function() {
        this.$data.config = this.$el.dataset;
        const rawContent = this.$el.innerText;
        this.$data.mentions = null;
        if (rawContent) {
            this.$data.mentions = JSON.parse(rawContent);
        }
    },
    render: function (createElement) {
        return createElement('Widget', {
            props: {
                title: this.$data.config.title || 'Mentions',
                endpoint: this.$data.config.endpoint,
                target: this.$data.config.target,
                showRSVPSummary: this.$data.config.rsvpSummary === 'yes',
                mentions: this.$data.mentions || null,
            }
        });
    }
});
