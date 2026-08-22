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

window.fkCombobox = {
    focusFromFace(evt) {
        if (evt.target.closest("button")) {
            return;
        }
        const input = evt.currentTarget.querySelector(".fk-combobox-input");
        if (input && evt.target !== input) {
            input.focus();
        }
    },
    renderChips(el, chips) {
        el.replaceChildren();
        for (const c of chips || []) {
            const chip = document.createElement("span");
            chip.className = "fk-chip";
            chip.append(document.createTextNode(c.label ?? ""));
            if (el.dataset.editable === "true") {
                const btn = document.createElement("button");
                btn.type = "button";
                btn.className = "fk-clear";
                btn.setAttribute("aria-label", "Remove");
                btn.dataset.id = String(c.id);
                btn.textContent = "×";
                btn.addEventListener("mousedown", (e) => e.preventDefault());
                chip.append(btn);
            }
            el.append(chip);
        }
    },
    keydown(evt, name) {
        const root = evt.currentTarget.closest(".fk-combobox");
        if (!root) {
            return;
        }
        const options = [...root.querySelectorAll(".fk-option")];
        if (evt.key === "Escape") {
            evt.preventDefault();
            return;
        }
        if (evt.key !== "ArrowDown" && evt.key !== "ArrowUp" && evt.key !== "Enter") {
            return;
        }
        evt.preventDefault();
        if (options.length === 0) {
            return;
        }
        let i = options.findIndex((el) => el.classList.contains("is-active"));
        if (evt.key === "ArrowDown") {
            i = i < 0 ? 0 : Math.min(i + 1, options.length - 1);
        } else if (evt.key === "ArrowUp") {
            i = i < 0 ? options.length - 1 : Math.max(i - 1, 0);
        } else if (evt.key === "Enter") {
            if (i >= 0) {
                options[i].click();
            }
            return;
        }
        options.forEach((el, j) => el.classList.toggle("is-active", j === i));
        options[i]?.scrollIntoView({ block: "nearest" });
    },
};

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
