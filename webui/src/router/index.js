import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import ChatView from '../views/ChatView.vue'



const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        { path: '/login', component: LoginView },
        { path: '/', component: HomeView, 
			children: [
        { path: 'conversations/:id', component: ChatView }
    				  ]
		},      
			]
})

// navigation guard: redirect to login if no token is found
router.beforeEach((to, from, next) => {
  const publicPages = ['/login'];
  const authRequired = !publicPages.includes(to.path);
  const loggedIn = localStorage.getItem('token');

  if (authRequired && !loggedIn) {
    return next('/login');
  }
  next();
});

export default router
