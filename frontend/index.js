import Vue from 'vue';
import App from './components/App.vue';
import Index from './components/Index.vue';
import Login from './components/Login.vue';
import Send from './components/Send.vue';
import Policies from './components/Policies.vue';
import Authenticate from './components/Authenticate.vue';
import Vuex from 'vuex';
import Router from 'vue-router';
import Axios from 'axios';
Vue.use(Vuex);
Vue.use(Router);

// Strip the "/ui/" from the current location:
const API_BASE_URL = window.location.pathname.substring(0, window.location.pathname.length - 4);

const transport = Axios.create({
  withCredentials: true
});

if (localStorage.getItem('session')) {
  transport.defaults.headers.common['Authorization'] = localStorage.getItem('session');
}

const store = new Vuex.Store({
  state: {
    loggedIn: !!localStorage.getItem('session'),
    authStatus: null,
    mentions: null,
    requestTokenStatus: null,
    getMentionsStatus: null,
    sendStatus: null,
    sendStatusReport: null,
    updateMentionStatusStatus: null,
    mentionFilterStatus: 'verified',
    policiesLoading: null,
    policiesError: null,
    policies: null,
    deletePolicyLoading: false,
    deletePolicyError: null,
    createPolicyLoading: false,
    createPolicyError: null
  },
  mutations: {
    logout(state) {
      state.loggedIn = false;
    },
    updateAuthStatus(state, status) {
      state.authStatus = status;
      if (status === 'succeeded') {
        state.loggedIn = true;
      }
    },
    updateRequestTokenStatus(state, newStatus) {
      state.requestTokenStatus = newStatus;
    },
    sendStatus(state, newStatus) {
      state.sendStatus = newStatus;
    },
    sendStatusReport(state, newStatus) {
      state.sendStatusReport = newStatus;
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
    setPoliciesLoading(state, val) {
      state.policiesLoading = val;
    },
    setPoliciesError(state, val) {
      state.policiesError = val;
    },
    setPolicies(state, val) {
      state.policies = val;
    },
    deletePolicyStarted(state) {
      state.deletePolicyLoading = true;
      state.deletePolicyError = null;
    },
    deletePolicySuccessful(state) {
      state.deletePolicyLoading = false;
      state.deletePolicyError = null;
    },
    deletePolicyFailed(state, e) {
      state.deletePolicyLoading = false;
      state.deletePolicyError = e;
    },
    createPolicyStarted(state) {
      state.createPolicyLoading = true;
      state.createPolicyError = null;
    },
    createPolicySuccessful(state) {
      state.createPolicyLoading = false;
      state.createPolicyError = null;
    },
    createPolicyFailed(state, e) {
      state.createPolicyLoading = false;
      state.createPolicyError = e;
    }
  },
  actions: {
    async authenticate(context, token) {
      context.commit('updateAuthStatus', 'pending');
      try {
        const data = new URLSearchParams();
        data.set('token', token);
        const resp = await transport.post(`${API_BASE_URL}/authenticate`, data);
        context.commit('updateAuthStatus', 'succeeded');
        localStorage.setItem('session', `Bearer ${resp.data}`);
        transport.defaults.headers.common['Authorization'] = localStorage.getItem('session');
      } catch(e) {
        console.log(e);
        context.commit('updateAuthStatus', 'failed');
      }
    },
    async sendMention(context, source) {
      context.commit('sendStatus', 'pending');
      context.commit('sendStatusReport', null);
      try {
        const resp = await transport.post(`${API_BASE_URL}/manage/send`, JSON.stringify({
          source,
        }));
        context.commit('sendStatus', 'succeeded');
        context.commit('sendStatusReport', resp.data);
      } catch (e) {
        if (e.response && e.response.status === 401) {
            context.commit('logout');
        }
        if (typeof e.response.data === 'object') {
          context.commit('sendStatusReport', e.response.data);
        }
        context.commit('sendStatus', 'failed');
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
        if (e.response && e.response.status === 401) {
            context.commit('logout');
        }
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
        console.log(e);
        if (e.response && e.response.status == 401) {
            context.commit('logout');
        }
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
        console.log(e);
        context.commit('updateMentionStatusStatus', 'failed');
      }
    },
    setMentionFilterStatus(context, status) {
      context.commit('setMentionFilterStatus', status);
    },
    logout(context) {
      localStorage.removeItem('session');
      context.commit('logout');
    },
    async getPolicies(context) {
      context.commit('setPoliciesLoading', true);
      context.commit('setPoliciesError', null);
      try {
        const resp = await transport.get(`${API_BASE_URL}/manage/policies`);
        context.commit('setPolicies', resp.data);
        context.commit('setPoliciesLoading', false);
      } catch(e) {
        context.commit('setPoliciesError', e);
        context.commit('setPoliciesLoading', false);
      }
    },
    async deletePolicy(context, id) {
      context.commit('deletePolicyStarted');
      try {
        await transport.delete(`${API_BASE_URL}/manage/policies/${id}`);
        context.commit('deletePolicySuccessful');
      } catch(e) {
        context.commit('deletePolicyFailed', e);
      }
    },
    async createPolicy(context, {urlPattern, weight, policy}) {
      context.commit('createPolicyStarted');
      try {
        await transport.post(`${API_BASE_URL}/manage/policies`, {
          url_pattern: urlPattern,
          weight,
          policy
        });
        context.commit('createPolicySuccessful');
      } catch(e) {
        context.commit('createPolicyFailed', e);
      }
    }
  }
});

const router = new Router({
  routes: [
    {
      path: '/authenticate/:token',
      component: Authenticate,
      meta: {
        title: 'Authenticate'
      },
    },
    {
      path: '/authenticate',
      component: Authenticate,
      meta: {
        title: 'Authenticate'
      },
    },
    {
      path: '/login',
      component: Login,
      meta: {
        title: 'Log in'
      },
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
      component: Index,
      meta: {
        title: 'Mentions'
      },
      beforeEnter: (to, from, next) => {
        if(!store.state.loggedIn) {
          next('/login');
          return;
        }
        next();
      }
    },
    {
      path: '/policies',
      component: Policies,
      meta: {
        title: 'Policies'
      },
      beforeEnter: (to, from, next) => {
        if(!store.state.loggedIn) {
          next('/login');
          return;
        }
        next();
      }
    },
    {
      path: '/send',
      component: Send,
      meta: {
        title: 'Send'
      },
      beforeEnter: (to, from, next) => {
        if(!store.state.loggedIn) {
          next('/login');
          return;
        }
        next();
      }
    },
  ]
});

router.beforeEach((to, from, next) => {
  const title = to.meta.title || 'Webmentions';
  document.querySelector('title').innerHTML = title;
  next();
});

new Vue({
  el: '#app',
  store,
  router,
  render: function(createElement) {
    return createElement(App);
  }
});
