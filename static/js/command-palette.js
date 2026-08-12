// Command palette helpers used from Datastar expressions.
window.commandPalette = {
    focusInput() {
        requestAnimationFrame(() => {
            document.getElementById("command-palette-input")?.focus();
        });
    },

    open(state) {
        state.query = "";
        state._open = true;
        state._index = 0;
    },

    close(state) {
        state._open = false;
    },

    isOpenShortcut(evt) {
        return (
            (evt.ctrlKey || evt.metaKey) &&
            !evt.shiftKey &&
            !evt.altKey &&
            evt.key.toLowerCase() === "p"
        );
    },

    isCloseShortcut(evt, state) {
        return evt.key === "Escape" && state._open;
    },

    onDialogKeydown(evt, state) {
        const items = document.querySelectorAll(
            "#command-palette-results .command-palette-item",
        );
        if (evt.key === "ArrowDown") {
            evt.preventDefault();
            if (items.length) {
                state._index = Math.min(state._index + 1, items.length - 1);
            }
            return;
        }
        if (evt.key === "ArrowUp") {
            evt.preventDefault();
            if (items.length) {
                state._index = Math.max(state._index - 1, 0);
            }
            return;
        }
        if (evt.key === "Enter") {
            evt.preventDefault();
            items[state._index]?.click();
        }
    },

    onWindowKeydown(evt, state) {
        if (this.isOpenShortcut(evt)) {
            evt.preventDefault();
            this.open(state);
            return;
        }
        if (this.isCloseShortcut(evt, state)) {
            evt.preventDefault();
            this.close(state);
        }
    },
};
