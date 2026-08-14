import {
    attribute,
    beginBatch,
    effect,
    endBatch,
    filtered,
    mergePatch,
    root,
} from "datastar";

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

// Subset of Datastar Pro's data-query-string: sync matching signals with
// the URL query string. Supports {include, exclude} filters and the
// __filter (omit empty values) and __history (pushState + popstate) modifiers.
attribute({
    name: "query-string",
    requirement: {
        key: "denied",
        value: "allowed",
    },
    returnsValue: true,
    apply({ rx, mods }) {
        const filterEmpty = mods.has("filter");
        const useHistory = mods.has("history");
        const signalFilter = normalizeSignalFilter(rx?.() ?? {});
        let applying = false;

        const applyFromURL = () => {
            applying = true;
            beginBatch();
            try {
                mergePatch(patchFromQuery(signalFilter));
            } finally {
                endBatch();
                applying = false;
            }
        };

        applyFromURL();

        const stop = effect(() => {
            JSON.stringify(root);
            if (applying) {
                return;
            }
            writeQuery(filtered(signalFilter), {
                filterEmpty,
                useHistory,
                signalFilter,
            });
        });

        if (!useHistory) {
            return stop;
        }

        const onPopState = () => applyFromURL();
        window.addEventListener("popstate", onPopState);
        return () => {
            stop();
            window.removeEventListener("popstate", onPopState);
        };
    },
});

function normalizeSignalFilter(raw) {
    return {
        include: raw?.include ?? /.*/,
        exclude: raw?.exclude ?? /(^_|\._)/,
    };
}

function matchesFilter(path, filter) {
    return filter.include.test(path) && !filter.exclude.test(path);
}

function flatten(value, prefix = "", out = []) {
    if (value !== null && typeof value === "object" && !Array.isArray(value)) {
        for (const [key, nested] of Object.entries(value)) {
            flatten(nested, prefix ? `${prefix}.${key}` : key, out);
        }
        return out;
    }
    if (prefix) {
        out.push([prefix, value]);
    }
    return out;
}

function setPath(target, path, value) {
    const parts = path.split(".");
    let cursor = target;
    for (let i = 0; i < parts.length - 1; i++) {
        const part = parts[i];
        if (cursor[part] == null || typeof cursor[part] !== "object") {
            cursor[part] = {};
        }
        cursor = cursor[part];
    }
    cursor[parts[parts.length - 1]] = value;
}

function isEmptyQueryValue(value) {
    return value == null || value === "";
}

function patchFromQuery(signalFilter) {
    const params = new URLSearchParams(window.location.search);
    const patch = {};
    for (const [path] of flatten(filtered(signalFilter))) {
        if (!matchesFilter(path, signalFilter)) {
            continue;
        }
        patchPath(patch, path, params.has(path) ? params.get(path) : "");
    }
    for (const [key, value] of params) {
        if (key === "datastar" || !matchesFilter(key, signalFilter)) {
            continue;
        }
        setPath(patch, key, value);
    }
    return patch;
}

function patchPath(target, path, value) {
    setPath(target, path, value);
}

function writeQuery(snapshot, { filterEmpty, useHistory, signalFilter }) {
    const params = new URLSearchParams(window.location.search);
    params.delete("datastar");
    for (const [path] of flatten(filtered(signalFilter))) {
        if (matchesFilter(path, signalFilter)) {
            params.delete(path);
        }
    }

    for (const [path, value] of flatten(snapshot)) {
        if (!matchesFilter(path, signalFilter)) {
            continue;
        }
        if (filterEmpty && isEmptyQueryValue(value)) {
            continue;
        }
        params.set(path, String(value));
    }

    const search = params.toString();
    const next = window.location.pathname + (search ? `?${search}` : "") + window.location.hash;
    const current = window.location.pathname + window.location.search + window.location.hash;
    if (next === current) {
        return;
    }
    if (useHistory) {
        history.pushState(null, "", next);
    } else {
        history.replaceState(null, "", next);
    }
}
