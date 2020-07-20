<template>
  <p class="paging">
    <a :class="previousLinkClass" href="#" @click="previousPage"><i class="fa fa-fw fa-arrow-left"></i> Previous page</a>
    <span class="paging__status">{{ offset + 1 }} - {{ offset + currentPageSize }} out of {{ total }}</span>
    <a :class="nextLinkClass" href="#" @click="nextPage">Next page <i class="fa fa-fw fa-arrow-right"></i></a>
  </p>
</template>
<script>
export default {
  props: {
    nextPage: Function,
    previousPage: Function,
    total: Number,
    offset: Number,
    currentPageSize: Number,
    pageSize: Number
  },
  computed: {
    nextLinkClass() {
      if (this.offset + this.currentPageSize >= this.total) {
        return "paging__next paging__next--disabled";
      }
      return "paging__next";
    },
    previousLinkClass() {
      if (this.offset === 0) {
        return "paging__previous paging__previous--disabled";
      }
      return "paging__previous";
    }
  }
}
</script>