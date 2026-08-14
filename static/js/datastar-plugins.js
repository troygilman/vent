import { attribute, filtered } from "datastar";

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

// data-push-url="{include: /foo/, exclude: /bar/}" writes matching signals
// into ?datastar= when a Datastar request on this element finishes.
window.addEventListener("popstate", () => location.reload());

attribute({
    name: "push-url",
    requirement: {
        key: "denied",
        value: "must",
    },
    returnsValue: true,
    apply({ el, rx }) {
        const onFetch = (event) => {
            if (event.detail?.type !== "finished") {
                return;
            }
            if (event.detail.el !== el && !el.contains(event.detail.el)) {
                return;
            }
            const filter = rx() || {};
            const signals = filtered({
                include: filter.include ?? /.*/,
                exclude: filter.exclude ?? /(^|\.)_/,
            });
            const params = new URLSearchParams();
            params.set("datastar", JSON.stringify(signals));
            const href = `${location.pathname}?${params}${location.hash}`;
            const current = location.pathname + location.search + location.hash;
            if (href !== current) {
                history.pushState({}, "", href);
            }
        };

        document.addEventListener("datastar-fetch", onFetch);
        return () => document.removeEventListener("datastar-fetch", onFetch);
    },
});
