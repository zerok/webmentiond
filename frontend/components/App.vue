<template>
  <div id="app">
    <router-view></router-view>
  </div>
</template>
<script>
import Login from './Login.vue';
export default {
  components: {
    Login,
  },
  computed: {
    loggedIn() {
      return this.$store.state.loggedIn;
    }
  },
  created() {
    console.log(this.$store.state.loggedIn);
  }
};
</script>
