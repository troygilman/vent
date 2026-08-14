import {
  attribute,
  effect,
  filtered,
  mergePaths,
  root,
} from "datastar"

/**
 * Confirmation dialog plugin.
 *
 * Usage:
 *   <button data-on:click="@confirm('Are you sure?') && @post('/delete')">
 */
attribute({
  name: "confirm",
  requirement: {
    key: "denied",
    value: "must",
  },
  argNames: ["message"],
  returnsValue: true,
  apply({ rx }) {
    return {
      confirm(message) {
        if (typeof message !== "string") {
          throw new Error("confirm: argument must be a string")
        }
        return window.confirm(message)
      },
    }
  },
})

/**
 * data-query-string plugin
 * Open-source approximation of the Datastar Pro attribute
 *
 * Public APIs used:
 *   - attribute()
 *   - effect()
 *   - filtered()
 *   - mergePaths()
 *   - root
 *
 * Usage:
 *   <div data-query-string></div>
 *   <div data-query-string="{include: /^(search|status|page)$/}"></div>
 *   <div data-query-string__filter></div>
 *   <div data-query-string__history></div>
 *   <div data-query-string__filter__history="{include: /^(search|page|sort)$/}"></div>
 *
 * After URL params are merged into signals, the element dispatches a bubbling
 * `query-string` event. Use data-on:query-string to fetch or otherwise react.
 */
attribute({
  name: "query-string",
  requirement: {
    key: "denied",
    value: "allowed",
  },
  returnsValue: true,

  apply({ el, mods, rx }) {
    const filterEmpty = mods.has("filter")
    const useHistory = mods.has("history")

    let filter = {}
    try {
      const result = rx?.()
      if (result && typeof result === "object") {
        filter = result
      }
    } catch {
      // no filter expression
    }

    const readUrl = () => {
      const params = new URLSearchParams(location.search)
      const obj = {}
      for (const [key, value] of params) {
        obj[key] = value
      }
      return obj
    }

    const includeRe = filter.include ? new RegExp(filter.include) : /.*/
    const excludeRe = filter.exclude ? new RegExp(filter.exclude) : /(?!)/

    const collectSnapshotPaths = (obj, prefix = "", paths = []) => {
      for (const [key, value] of Object.entries(obj)) {
        const path = prefix ? `${prefix}.${key}` : key
        if (value !== null && typeof value === "object" && !Array.isArray(value)) {
          collectSnapshotPaths(value, path, paths)
        } else if (includeRe.test(path) && !excludeRe.test(path)) {
          paths.push(path)
        }
      }
      return paths
    }

    const applyUrlToSignals = () => {
      const urlParams = readUrl()
      const paths = []
      const seen = new Set()

      for (const [path, raw] of Object.entries(urlParams)) {
        if (path === "datastar") continue
        if (!includeRe.test(path) || excludeRe.test(path)) continue
        seen.add(path)
        paths.push([path, raw])
      }

      for (const path of collectSnapshotPaths(filtered(filter, root))) {
        if (!seen.has(path)) {
          paths.push([path, ""])
        }
      }

      if (paths.length) {
        mergePaths(paths)
      }
    }

    const writeSignalsToUrl = (push = false) => {
      const snapshot = filtered(filter, root)
      const params = new URLSearchParams()

      const walk = (obj, prefix = "") => {
        for (const [key, value] of Object.entries(obj)) {
          const path = prefix ? `${prefix}.${key}` : key

          if (value !== null && typeof value === "object" && !Array.isArray(value)) {
            walk(value, path)
          } else {
            if (filterEmpty && (value === "" || value == null || value === undefined)) {
              continue
            }
            params.set(path, String(value))
          }
        }
      }

      walk(snapshot)

      const qs = params.toString()
      const newUrl =
        location.pathname +
        (qs ? `?${qs}` : "") +
        location.hash

      if (newUrl === location.pathname + location.search + location.hash) return

      if (push && useHistory) {
        history.pushState(null, "", newUrl)
      } else {
        history.replaceState(null, "", newUrl)
      }
    }

    const emitUpdated = () => {
      el.dispatchEvent(new Event("query-string", { bubbles: true }))
    }

    applyUrlToSignals()
    queueMicrotask(() => {
      applyUrlToSignals()
      emitUpdated()
    })

    const stopEffect = effect(() => {
      writeSignalsToUrl(useHistory)
    })

    let onPopState
    if (useHistory) {
      onPopState = () => {
        applyUrlToSignals()
        emitUpdated()
      }
      window.addEventListener("popstate", onPopState)
    }

    return () => {
      stopEffect()
      if (onPopState) {
        window.removeEventListener("popstate", onPopState)
      }
    }
  },
})
