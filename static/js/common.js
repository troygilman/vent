// CSRF Token Header Middleware
(function () {
    const originalFetch = window.fetch.bind(window);

    window.fetch = function (input, init) {
        init = init || {};
        const method = (init.method || "GET").toUpperCase();
        if (method !== "GET" && method !== "HEAD") {
            const token = document
                .querySelector('meta[name="csrf-token"]')
                ?.getAttribute("content");
            if (token) {
                const headers = new Headers(init.headers || {});
                headers.set("X-CSRF-Token", token);
                init = Object.assign({}, init, { headers: headers });
            }
        }
        return originalFetch(input, init);
    };
})();

window.widgetDrawer = {
    toggle(state, name) {
        if (state._open && state.active === name) {
            state._open = false;
            return;
        }
        state.active = name;
        state._open = true;
    },
};
