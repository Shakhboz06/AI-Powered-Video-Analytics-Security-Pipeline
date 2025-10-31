<template>
  <admin-layout>
    <div class="grid grid-cols-12 gap-4 md:gap-6">
      <div class="col-span-12 space-y-6 xl:col-span-7">
        <ecommerce-metrics />
        <monthly-target />
      </div>
      <div class="col-span-12 xl:col-span-5">
        <monthly-sale />
      </div>

      <div class="col-span-12">
        <statistics-chart />
      </div>

      <div class="col-span-12 xl:col-span-5">
        <customer-demographic />
      </div>

      <div class="col-span-12 xl:col-span-7">
        <recent-orders />
      </div>
    </div>
  </admin-layout>
</template>

<script>
import AdminLayout from '../app/components/layout/AdminLayout.vue'
import EcommerceMetrics from '../app/components/ecommerce/EcommerceMetrics.vue'
import MonthlyTarget from '../app/components/ecommerce/MonthlySale.vue'
import MonthlySale from '../app/components/ecommerce/MonthlyTarget.vue'
import CustomerDemographic from '../app/components/ecommerce/CustomerDemographic.vue'
import StatisticsChart from '../app/components/ecommerce/StatisticsChart.vue'
import RecentOrders from '../app/components/ecommerce/RecentOrders.vue'
export default {
  components: {
    AdminLayout,
    EcommerceMetrics,
    MonthlyTarget,
    MonthlySale,
    CustomerDemographic,
    StatisticsChart,
    RecentOrders,
  },
  name: 'Ecommerce',
}
</script>


<!-- <template>
  <div>
    sadsafdsfdsfsdfdfgd
    <h1>adadadadasd</h1>
  </div>
</template>

<script lang="ts" setup>

</script>

<style>

</style> -->