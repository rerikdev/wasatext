import { createRouter, createWebHashHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import ProfileView from '../views/ProfileView.vue'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', component: LoginView },
    { path: '/home', component: HomeView },
    { path: '/profile', component: ProfileView },
  ]
})

// Navigation guard globale
router.beforeEach((to, from, next) => {
  const isLoggedIn = !!localStorage.getItem('userId');
  if (!isLoggedIn && to.path !== '/') {
    // Se non loggato, manda sempre al login
    next('/');
  } else if (isLoggedIn && to.path === '/') {
    // Se già loggato, evita di tornare al login
    next('/home');
  } else {
    next();
  }
});

export default router
