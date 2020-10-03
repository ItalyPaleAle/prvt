{#if $modal}
    <Modal />
{/if}

<div class="container w-full lg:w-3/5 px-2 pt-6 lg:pt-10">
    <Router {routes} on:routeLoaded={routeLoaded} />

    <Footer />
</div>

<svelte:body on:click={bodyClick} />

<script>
import Router from 'svelte-spa-router'

// Routes
import routes from './routes'

// Components
import Modal from '../shared/components/Modal.svelte'
import Footer from '../shared/components/Footer.svelte'

// Stores
import {modal} from '../shared/stores'

// Clicking on the background anywhere will hide any modal currently open
function bodyClick(event) {
    // Only capture clicks on the body, and not child elements
    if (event && event.target == document.body && !event.defaultPrevented) {
        $modal = null
    }
}

// When the page is changed, hide the modal
function routeLoaded(event) {
    $modal = null
}
</script>
