<template>
  <form v-on:submit="onSubmit">
    <label>Token: <input v-model="token"></label>
    <button type="submit">Authenticate</button>
  </form>
</template>
<script>
  export default {
    methods: {
      onSubmit(evt) {
        this.$store.dispatch('authenticate', this.$data.token);
      }
    },
    data: function() {
      return {
        token: this.$route.params.token
      };
    }
  };
</script>
