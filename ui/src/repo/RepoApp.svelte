{#if $modal}
    <Modal />
{/if}

<div class="container w-full lg:w-3/5 px-2 pt-6 lg:pt-10">
    <div class="flex justify-between items-center w-full mb-10 lg:mb-12">
        <h1 class="font-bold break-normal text-text-300 px-2 text-2xl md:text-3xl truncate">
            prvt repository
        </h1>
    </div>

    <Router {routes} />

    <Footer />
</div>

<svelte:body on:click={bodyClick} />

<script>
// Libraries
import Router from 'svelte-spa-router'

// Components
import Modal from '../shared/components/Modal.svelte'
import Footer from '../shared/components/Footer.svelte'

// Routes
import routes from './routes'

// Stores
import {modal} from '../shared/stores'

// Clicking on the background anywhere will hide any modal currently open
function bodyClick(event) {
    // Only capture clicks on the body, and not child elements
    if (event && event.target == document.body && !event.defaultPrevented) {
        $modal = null
    }
}
</script>
