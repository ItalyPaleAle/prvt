<Navbar />

<div class="container w-full lg:w-3/5 px-2 pt-4 lg:pt-12 mt-10">
    <Router {routes} />

    <Footer />
</div>

<svelte:body on:click={bodyClick} />

<script>
import Router from 'svelte-spa-router'
import active from 'svelte-spa-router/active'

// Routes
import routes from './routes'

// Components
import Navbar from './components/Navbar.svelte'
import Footer from './components/Footer.svelte'

// Stores
import {dropdown} from './stores'

// Clicking on the background anywhere will hide any dropdown menu currently open
function bodyClick(event) {
    // Only capture clicks on the body, and not child elements
    if (event && event.target == document.body && !event.defaultPrevented) {
        $dropdown = null
    }
}
</script>
