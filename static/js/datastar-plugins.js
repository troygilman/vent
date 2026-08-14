import { attribute } from "datastar";

attribute({
    name: "confirm",
    requirement: {
        key: "optional",
        value: "must",
    },
    returnsValue: true,
    apply({ el, rx, key }) {
        const eventName = key || "click";

        const handler = (e) => {
            const message = rx();
            if (!confirm(message)) {
                e.preventDefault();
                e.stopImmediatePropagation();
            }
        };

        el.addEventListener(eventName, handler, true);

        return () => {
            el.removeEventListener(eventName, handler, true);
        };
    },
});

// data-replace-url writes the last Datastar GET URL into history when a
// request finishes on this element. Default is pushState; __replace uses
// replaceState. Back/forward reload so the server re-renders from ?datastar=.
const fetchURLs = new WeakMap();
let pendingInitiator = null;
let fetchPatched = false;
let popstateBound = false;

function requestURL(input) {
    if (typeof input === "string") {
        return input;
    }
    if (input instanceof URL) {
        return input.href;
    }
    if (input && typeof input.url === "string") {
        return input.url;
    }
    return "";
}

function isDatastarGET(input, init) {
    const method = (init?.method || (input && input.method) || "GET").toUpperCase();
    if (method !== "GET") {
        return false;
    }
    const headers = new Headers(init?.headers || (input && input.headers) || {});
    return headers.get("Datastar-Request") === "true";
}

function historyHref(url) {
    const parsed = new URL(url, window.location.href);
    return parsed.pathname + parsed.search + parsed.hash;
}

function patchFetch() {
    if (fetchPatched) {
        return;
    }
    fetchPatched = true;

    const originalFetch = window.fetch.bind(window);
    window.fetch = function (input, init) {
        if (pendingInitiator && isDatastarGET(input, init)) {
            fetchURLs.set(pendingInitiator, requestURL(input));
            pendingInitiator = null;
        }
        return originalFetch(input, init);
    };

    document.addEventListener("datastar-fetch", (event) => {
        if (event.detail?.type === "started") {
            pendingInitiator = event.detail.el;
        }
    });
}

function bindPopstateReload() {
    if (popstateBound) {
        return;
    }
    popstateBound = true;
    window.addEventListener("popstate", () => {
        window.location.reload();
    });
}

attribute({
    name: "replace-url",
    requirement: {
        key: "denied",
        value: "denied",
    },
    apply({ el, mods }) {
        patchFetch();
        const replace = mods.has("replace");
        if (!replace) {
            bindPopstateReload();
        }

        const onFetch = (event) => {
            if (event.detail?.type !== "finished") {
                return;
            }
            const initiator = event.detail.el;
            if (initiator !== el && !el.contains(initiator)) {
                return;
            }
            const url = fetchURLs.get(initiator);
            if (!url) {
                return;
            }
            const next = historyHref(url);
            const current = window.location.pathname + window.location.search + window.location.hash;
            if (next === current) {
                return;
            }
            if (replace) {
                history.replaceState(null, "", next);
            } else {
                history.pushState(null, "", next);
            }
        };

        document.addEventListener("datastar-fetch", onFetch);
        return () => document.removeEventListener("datastar-fetch", onFetch);
    },
});
