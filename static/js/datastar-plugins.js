import { action, attribute, filtered } from "datastar";

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

// @pushHistory(regex) writes matching signals into ?datastar= and pushStates.
// Use the same include regex as @get's filterSignals so the bar matches the request.
window.addEventListener("popstate", () => location.reload());

action({
    name: "pushHistory",
    apply(_ctx, include) {
        const signals = filtered({ include, exclude: /(^|\.)_/ });
        const params = new URLSearchParams();
        params.set("datastar", JSON.stringify(signals));
        const href = `${location.pathname}?${params}${location.hash}`;
        const current = location.pathname + location.search + location.hash;
        if (href !== current) {
            history.pushState({}, "", href);
        }
    },
});
