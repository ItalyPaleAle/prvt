{#if $modal}
    <Modal />
{/if}

<div class="container w-full lg:w-3/5 px-2 pt-6 lg:pt-10">
    {#if !hide}
        <Router {routes} on:routeLoaded={routeLoaded} />
    {/if}

    <Footer />
</div>

<svelte:body on:click={bodyClick} />

<script lang="ts">
import Router from 'svelte-spa-router'
import type {RouterEvent, RouteDetailLoaded} from 'svelte-spa-router'

// Props
export let hide = false

// Routes
import routes from './routes'

// Components
import Modal from './components/Modal.svelte'
import Footer from './components/Footer.svelte'

// Stores
import {modal} from './stores'

// Clicking on the background anywhere will hide any modal currently open
function bodyClick(event: MouseEvent) {
    // Only capture clicks on the body, and not child elements
    if (event?.target == document.body && !event?.defaultPrevented) {
        $modal = null
    }
}

// When the page is changed, hide the modal
function routeLoaded(_: RouterEvent<RouteDetailLoaded>) {
    $modal = null
}
</script>
