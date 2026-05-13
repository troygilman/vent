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
