// Adapted from https://github.com/PolymerElements/platinum-sw/blob/31465a989a53aed1d63249ae7d9bb5a1268cb6b6/test/controlled-promise.js

/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at
http://polymer.github.io/LICENSE.txt The complete set of authors may be found at
http://polymer.github.io/AUTHORS.txt The complete set of contributors may be
found at http://polymer.github.io/CONTRIBUTORS.txt Code distributed by Google as
part of the polymer project is also subject to an additional IP rights grant
found at http://polymer.github.io/PATENTS.txt
*/
// Provides an equivalent to navigator.serviceWorker.ready that waits for the
// page to be controlled, as opposed to waiting for the active service worker.
// See https://github.com/slightlyoff/ServiceWorker/issues/799
export default new Promise<ServiceWorkerRegistration>((resolve, reject) => {
    // Resolve with the registration, to match the .ready promise's behavior.
    const resolveWithRegistration = () => {
        navigator.serviceWorker
            .getRegistration()
            .then((registration) => {
                if (registration) {
                    resolve(registration)
                }
                else {
                    reject(new Error('Empty registration received'))
                }
            })
    }
  
    if (navigator.serviceWorker.controller) {
        resolveWithRegistration()
    }
    else {
        navigator.serviceWorker.addEventListener('controllerchange', resolveWithRegistration)
    }
})
