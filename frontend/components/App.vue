<template>
  <div id="app" :class="'app app--' + (loggedIn ? 'logged-in' : 'logged-out')">
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
  computed: {
    loggedIn() {
      return this.$store.state.loggedIn;
    },
    ...mapState(['loggedIn'])
  },
  updated() {
    if (!this.loggedIn && this.$route.path !== '/login') {
      this.$router.push('/login');
    }
  },
  created() {
    console.log(this.$store.state.loggedIn);
  }
};
</script>
