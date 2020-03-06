<Navbar />

<div class="container w-full lg:w-3/5 px-2 pt-8 lg:pt-16 mt-16">
    <Router {routes} />

    <Footer />
</div>

<script>
import Router from 'svelte-spa-router'
import active from 'svelte-spa-router/active'

// Routes
import routes from './routes'

// Components
import Navbar from './components/Navbar.svelte'
import Footer from './components/Footer.svelte'
</script>
