<template>
  <div id="app" :class="'app app--' + (loggedIn ? 'logged-in' : 'logged-out')">
    <a v-if="loggedIn" href="#" v-on:click="logout">Log out</a>
    <router-view></router-view>
  </div>
</template>
<script>
  import {mapState} from 'vuex';
import Login from './Login.vue';
export default {
  components: {
    Login,
  },
  methods: {
    logout(evt) {
      evt.preventDefault();
      this.$store.dispatch('logout');
    }
  },
  computed: {
    loggedIn() {
      return this.$store.state.loggedIn;
    },
    ...mapState(['loggedIn'])
  },
  updated() {
    if (!this.loggedIn && !(this.$route.path === "/login" || this.$route.path === '/authenticate')) {
      this.$router.push('/login');
    }
  },
};
</script>
