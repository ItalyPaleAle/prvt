<footer class="block p-3 mt-8 mb-2 leading-normal text-xs">
    <p>Powered by <a href="https://github.com/ItalyPaleAle/prvt">prvt</a> {appVersion}</p>
    <ThemePicker />
</footer>

<script>
import ThemePicker from './ThemePicker.svelte'
const appVersion = env.APP_VERSION
</script>
