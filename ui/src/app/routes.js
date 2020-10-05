import {push} from 'svelte-spa-router'
import {wrap} from 'svelte-spa-router/wrap'

import AppInfo from './lib/appinfo'

// Components
import Tree from './routes/Tree.svelte'
import Add from './routes/Add.svelte'
import View from './routes/View.svelte'
import NotFound from './routes/NotFound.svelte'
import ConnectionList from './routes/ConnectionList.svelte'
import UnlockRepo from './routes/UnlockRepo.svelte'

// Route definition object
export default {
    // Tree
    '/': wrap({
        component: Tree,
        conditions: [
            requireUnlocked
        ]
    }),
    '/tree': wrap({
        component: Tree,
        conditions: [
            requireUnlocked
        ]
    }),
    '/tree/*': wrap({
        component: Tree,
        conditions: [
            requireUnlocked
        ]
    }),

    // Add
    '/add/*': wrap({
        component: Add,
        conditions: [
            requireUnlocked,
            noReadOnly
        ]
    }),

    // View
    '/view/:fileId': wrap({
        component: View,
        conditions: [
            requireUnlocked
        ]
    }),

    // Repo select
    '/repo': ConnectionList,

    // Unlock repo
    '/unlock': wrap({
        component: UnlockRepo,
        conditions: [
            requireRepo
        ]
    }),
        
    // Catch-all, must be last
    '*': NotFound,
}

// Allow a route only if a repo is selected (but not necessarily unlocked);
// otherwise, redirects to /repo to select a repo
async function requireRepo() {
    const info = await AppInfo.get()
    if (!info || !info.repoSelected) {
        push('/repo')
        return false
    }
    return true
}

// Allow a route only if the repo is unlocked;
// otherwise, redirects to /repo to select a repo
async function requireUnlocked() {
    const info = await AppInfo.get()
    if (!info || !info.repoUnlocked) {
        push('/repo')
        return false
    }
    return true
}

// Allow a route only in non read-only mode
async function noReadOnly() {
    const isReadOnly = await AppInfo.isReadOnly()
    return !isReadOnly
}
