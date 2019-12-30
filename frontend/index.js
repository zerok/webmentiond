import Vue from 'vue';
import App from './components/App.vue';
import Index from './components/Index.vue';
import Login from './components/Login.vue';
import Authenticate from './components/Authenticate.vue';
import Vuex from 'vuex';
import Router from 'vue-router';
import Axios from 'axios';
import Cookie from 'js-cookie';
Vue.use(Vuex);
Vue.use(Router);

const API_BASE_URL = "http://localhost:8080"

const transport = Axios.create({
  withCredentials: true
});

const store = new Vuex.Store({
  state: {
    loggedIn: !!Cookie.get('token'),
    authStatus: null,
    mentions: null,
    requestTokenStatus: null,
    getMentionsStatus: null,
    updateMentionStatusStatus: null,
    mentionFilterStatus: 'verified',
  },
  mutations: {
    updateAuthStatus(state, status) {
      state.authStatus = status;
    },
    updateRequestTokenStatus(state, newStatus) {
      state.requestTokenStatus = newStatus;
    },
    updateGetMentionsStatus(state, newStatus) {
      state.getMentionsStatus = newStatus;
    },
    updateMentionStatusStatus(state, newStatus) {
      state.updateMentionStatusStatus = newStatus;
    },
    setMentions(state, items) {
      state.mentions = items;
    },
    setMentionFilterStatus(state, status) {
      state.mentionFilterStatus = status;
    },
  },
  actions: {
    async authenticate(context, token) {
      context.commit('updateAuthStatus', 'pending');
      try {
        const data = new URLSearchParams();
        data.set('token', token);
        const resp = await transport.post(`${API_BASE_URL}/authenticate`, data);
        context.commit('updateAuthStatus', 'succeeded');
        console.log(resp);
      } catch(e) {
        console.log(e);
        context.commit('updateAuthStatus', 'failed');
      }
    },
    async getMentions(context, filter) {
      const f = Object.assign({
        status: 'verified'
      }, filter)
      context.commit('updateGetMentionsStatus', 'pending');
      try {
        context.commit('updateGetMentionsStatus', 'succeeded');
        const resp = await transport.get(`${API_BASE_URL}/manage/mentions?status=${f.status}`);
        context.commit('setMentions', resp.data.items);
      } catch (e) {
        console.log(e);
        context.commit('updateGetMentionsStatus', 'failed');
      }
    },
    async requestToken(context, email) {
      context.commit('updateRequestTokenStatus', 'pending');
      try {
        const data = new URLSearchParams();
        data.set('email', email);
        await transport.post(`${API_BASE_URL}/request-login`, data);
        context.commit('updateRequestTokenStatus', 'succeeded');
      } catch(e) {
        context.commit('updateRequestTokenStatus', 'failed');
      }
    },
    async approveMention(context, mention) {
      context.commit('updateMentionStatusStatus', 'pending');
      try {
        await transport.post(`${API_BASE_URL}/manage/mentions/${mention.id}/approve`);
        context.commit('updateMentionStatusStatus', 'succeeded');
      } catch(e) {
        context.commit('updateMentionStatusStatus', 'failed');
      }
    },
    async rejectMention(context, mention) {
      context.commit('updateMentionStatusStatus', 'pending');
      try {
        await transport.post(`${API_BASE_URL}/manage/mentions/${mention.id}/reject`);
        context.commit('updateMentionStatusStatus', 'succeeded');
      } catch(e) {
        context.commit('updateMentionStatusStatus', 'failed');
      }
    },
    setMentionFilterStatus(context, status) {
      context.commit('setMentionFilterStatus', status);
    },
  }
});

const router = new Router({
  routes: [
    {
      path: '/authenticate/:token',
      component: Authenticate
    },
    {
      path: '/login',
      component: Login,
      beforeEnter: (to, from, next) => {
        if(store.state.loggedIn) {
          next('/');
          return;
        }
        next();
      }
    },
    {
      path: '/',
      component: Index
    },
  ]
});

new Vue({
  el: '#app',
  store,
  router,
  render: function(createElement) {
    return createElement(App);
  }
});
