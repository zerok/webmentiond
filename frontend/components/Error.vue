<template>
  <div class="error" v-if="err">
    <i class="fa fa-exclamation-triangle"></i> {{ err }}
  </div>
</template>
<script>
export default {
  props: ['err']
}
</script>
