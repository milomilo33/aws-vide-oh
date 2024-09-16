<template>
  <div>
    <b-navbar fixed="top" toggleable="lg" type="light" variant="light">
      <b-navbar-brand :to="{ path: '/' }">Home Page</b-navbar-brand>
      <b-navbar-toggle target="nav-collapse"></b-navbar-toggle>
      <b-collapse id="nav-collapse" is-nav>    
        <b-navbar-nav>
          <!-- Other navigation items can go here if needed -->
        </b-navbar-nav>  

        <!-- Right aligned nav items -->
        <b-navbar-nav class="ml-auto">
          <b-nav-item :to="{ path: '/Login' }">Login</b-nav-item>
          <b-nav-item :to="{ path: '/Register' }">Register</b-nav-item>
        </b-navbar-nav>
      </b-collapse>
    </b-navbar>

    <router-view></router-view>
  </div>
</template>
