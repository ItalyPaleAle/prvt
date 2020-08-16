{#if $modal}
    <Modal />
{/if}
<Navbar />

<div class="container w-full lg:w-3/5 px-2 pt-4 lg:pt-12 mt-10">
    <Router {routes} on:routeLoaded={routeLoaded} />

    <Footer />
</div>

<svelte:body on:click={bodyClick} />

<script>
import Router from 'svelte-spa-router'
import active from 'svelte-spa-router/active'

// Routes
import routes from './routes'

// Components
import Modal from './components/Modal.svelte'
import Navbar from './components/Navbar.svelte'
import Footer from './components/Footer.svelte'

// Stores
import {dropdown, modal} from './stores'

// Clicking on the background anywhere will hide any dropdown menu or modal currently open
function bodyClick(event) {
    // Only capture clicks on the body, and not child elements
    if (event && event.target == document.body && !event.defaultPrevented) {
        $dropdown = null
        $modal = null
    }
}

// When a route is changed, hide all dropdowns or modals
function routeLoaded(event) {
    $dropdown = null
    $modal = null
}
</script>
