<h1>Test wasm app</h1>

<button on:click={loadImageFetch} class="border-black border p-2 m-2">
    Load image via fetch
</button>
<button on:click={() => imageEmbed = reset() && 'http://localhost:3129/rawfile/f0fc00d8-a369-4b42-9037-9613288d30b3'} class="border-black border p-2 m-2">
    Load image via embed
</button>
<button on:click={loadText} class="border-black border p-2 m-2">
    Load text
</button>
<button on:click={() => videoEmbed = reset() && 'http://localhost:3129/rawfile/84d4c1d2-e774-4eff-96e3-2314f72bb6eb'} class="border-black border p-2 m-2">
    Load video via embed
</button>

{#if outImageFetch}
    <img src="{outImageFetch}" alt="Decrypted" class="w-auto h-auto max-w-full max-h-90vh mx-auto" />
{/if}
{#if imageEmbed}
    <img src="{imageEmbed}" alt="Decrypted" class="w-auto h-auto max-w-full max-h-90vh mx-auto" />
{/if}
{#if outText}
    <div class="w-auto h-auto max-w-full max-h-90vh mx-auto text-xs overflow-scroll">
        <pre>{outText}</pre>
    </div>
{/if}
{#if videoEmbed}
    <video controls class="w-auto h-auto max-w-full max-h-90vh mx-auto">
        <source src={videoEmbed} type="video/x-m4v" />
    </video>
{/if}

<script>
let outText, outImageFetch, imageEmbed, videoEmbed

function reset() {
    outText = ''
    imageEmbed = null
    videoEmbed = null
    if (outImageFetch) {
        URL.revokeObjectURL(outImageFetch)
        outImageFetch = null
    }
    return true
}

async function loadText() {
    reset()
    const decoder = new TextDecoder('utf-8')
    const res = await fetch('http://localhost:3129/rawfile/015d4c16-2d95-4059-9f99-d91055c7a955')
    const reader = res.body.getReader()
    let done = false
    while (!done) {
        const read = await reader.read()
        done = read && read.done
        outText += decoder.decode(read.value)
    }
}

async function loadImageFetch() {
    reset()
    const res = await fetch('http://localhost:3129/rawfile/6dcb358b-c777-4eaa-95f5-ec89fb1493b9')
    const blob = await res.blob()
    outImageFetch = URL.createObjectURL(blob)
}
</script>
