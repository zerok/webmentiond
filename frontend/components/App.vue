<template>
  <div id="app" :class="'app app--' + (loggedIn ? 'logged-in' : 'logged-out')">
    <header class="topbar">
      <nav class="topnav" v-if="loggedIn">
        <div class="topnav__left">
          <a href="#/">Mentions</a>
          <a href="#/send">Send</a>
        </div>
        <div class="topnav__right">
          <a href="#" v-on:click="logout">Log out <i class="fa fa-sign-out"></i></a>
        </div>
      </nav>
    </header>
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
