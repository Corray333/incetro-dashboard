<script setup lang="ts">
import { onBeforeMount, ref } from 'vue';
import { DashboardTransport } from './transport/dashboard';
declare const Telegram: any

const authorized = ref(false)

onBeforeMount(async()=>{
  try {
    await DashboardTransport.authorize()
    authorized.value = true

    const tg = Telegram.WebApp

    tg.expand()
    tg.isVerticalSwipesEnabled = false
    tg.disableVerticalSwipes()
    tg.ready()

  } catch (error) {
    console.log(error)
  }
})

</script>

<template>
  <RouterView v-if="authorized"/>
  <div v-else class="text-center w-screen h-screen flex items-center justify-center font-bold">
    <p>Сори, я очень устал, пока делал авторизацию,<br>поэтому тут пока просто будет этот текст😐</p>
  </div>
</template>

<style scoped>



</style>
