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
    toggleOpen(state) {
        if (state._open) {
            state._open = false;
            return;
        }
        if (!state.active) {
            state.active = "filter";
        }
        state._open = true;
    },
    open(state, name) {
        state.active = name;
        state._open = true;
    },
};

window.tableFilters = {
    dispatch(el, detail) {
        const form =
            el.closest("#schema-table-filters") || el.closest("form") || el;
        form.dispatchEvent(
            new CustomEvent("table-filter-reset", {
                bubbles: true,
                detail: detail || {},
            }),
        );
    },
    onReset(filter, evt) {
        const names = (evt.detail || {}).filterNames;
        if (!Array.isArray(names) || !names.length) {
            return;
        }
        for (const name of names) {
            filter[name] = "";
        }
        evt.currentTarget.dispatchEvent(new Event("change"));
    },
};

window.tableFetch = {
    dispatch(el) {
        const form =
            el.closest("#schema-table-filters") || el.closest("form") || el;
        window.tableFetch.resetScroll();
        form.dispatchEvent(new Event("table-fetch"));
    },
    resetScroll() {
        const box = document.querySelector(".schema-table .table-container");
        if (!box) {
            return;
        }
        box.scrollTop = 0;
        box.scrollLeft = 0;
        const fresh = box.cloneNode(false);
        while (box.firstChild) {
            fresh.appendChild(box.firstChild);
        }
        box.replaceWith(fresh);
        fresh.scrollTop = 0;
        fresh.scrollLeft = 0;
    },
};

(function () {
    const queueTableScrollReset = () => {
        const apply = () => window.tableFetch.resetScroll();
        apply();
        queueMicrotask(apply);
        requestAnimationFrame(() => {
            apply();
            requestAnimationFrame(apply);
        });
    };

    document.addEventListener("datastar-fetch", (evt) => {
        const detail = evt.detail || {};
        const el = detail.el;
        if (!el || typeof el.closest !== "function") {
            return;
        }
        if (!el.closest("#schema-table-filters")) {
            return;
        }
        if (
            detail.type !== "finished" &&
            detail.type !== "datastar-patch-elements"
        ) {
            return;
        }
        queueTableScrollReset();
    });
})();
