export default defineNuxtRouteMiddleware((to, from) => {
  if (process.client) {
    const token = localStorage.getItem('admin_token')
    if (!token) {
      return navigateTo('/login')
    }
  }
})