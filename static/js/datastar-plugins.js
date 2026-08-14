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

// Datastar's datastar-fetch event is {type, el, argsRaw} — no URL.
// Record the GET URL from fetch, then push it on finished.
let pendingEl = null;
const urls = new WeakMap();
let installed = false;

function install() {
    if (installed) {
        return;
    }
    installed = true;

    const originalFetch = window.fetch.bind(window);
    window.fetch = (input, init) => {
        if (pendingEl) {
            urls.set(pendingEl, typeof input === "string" ? input : "");
            pendingEl = null;
        }
        return originalFetch(input, init);
    };

    document.addEventListener("datastar-fetch", (event) => {
        if (event.detail?.type === "started") {
            pendingEl = event.detail.el;
        }
    });

    window.addEventListener("popstate", () => location.reload());
}

attribute({
    name: "replace-url",
    requirement: {
        key: "denied",
        value: "denied",
    },
    apply({ el }) {
        install();

        const onFinished = (event) => {
            if (event.detail?.type !== "finished") {
                return;
            }
            if (event.detail.el !== el && !el.contains(event.detail.el)) {
                return;
            }
            const url = urls.get(event.detail.el);
            if (!url) {
                return;
            }
            const next = new URL(url, location.href);
            const href = next.pathname + next.search + next.hash;
            const current = location.pathname + location.search + location.hash;
            if (href !== current) {
                history.pushState(null, "", href);
            }
        };

        document.addEventListener("datastar-fetch", onFinished);
        return () => document.removeEventListener("datastar-fetch", onFinished);
    },
});
