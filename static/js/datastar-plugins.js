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
 * Falsey signal values are omitted when writing the query string.
 */
attribute({
  name: "query-string",
  requirement: {
    key: "denied",
    value: "allowed",
  },
  returnsValue: true,

  apply({ el, mods, rx }) {
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
          } else if (value) {
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

/**
 * data-cookie plugin
 *
 * Persists filtered signals to a browser-readable cookie (not HttpOnly) and
 * restores them on load. Mirrors data-query-string / Datastar Pro data-persist
 * filtering, but uses a cookie so state survives full page navigations.
 *
 * Usage:
 *   <div data-cookie></div>
 *   <div data-cookie="{include: /^widgets\./}"></div>
 *   <div data-cookie:widgets="{include: /^widgets\./, path: '/admin/'}"></div>
 *
 * The optional key becomes the cookie name with a vent- prefix
 * (data-cookie:widgets → vent-widgets). Default name is vent-datastar.
 * path and maxAge on the expression set cookie Path and Max-Age.
 */
attribute({
  name: "cookie",
  requirement: {
    key: "allowed",
    value: "allowed",
  },
  returnsValue: true,

  apply({ key, rx }) {
    let filter = {}
    try {
      const result = rx?.()
      if (result && typeof result === "object") {
        filter = result
      }
    } catch {
      // no filter expression
    }

    const cookieName = key ? `vent-${key}` : "vent-datastar"
    const cookiePath = typeof filter.path === "string" && filter.path ? filter.path : "/"
    const maxAge =
      Number.isFinite(filter.maxAge) && filter.maxAge > 0
        ? filter.maxAge
        : 365 * 24 * 60 * 60
    const signalFilter = {}
    if (filter.include) signalFilter.include = filter.include
    if (filter.exclude) signalFilter.exclude = filter.exclude
    const includeRe = filter.include ? new RegExp(filter.include) : /.*/
    const excludeRe = filter.exclude ? new RegExp(filter.exclude) : /(?!)/

    const flatten = (obj, prefix = "", paths = []) => {
      if (obj === null || typeof obj !== "object" || Array.isArray(obj)) {
        if (prefix && includeRe.test(prefix) && !excludeRe.test(prefix)) {
          paths.push([prefix, obj])
        }
        return paths
      }
      for (const [name, value] of Object.entries(obj)) {
        const path = prefix ? `${prefix}.${name}` : name
        if (value !== null && typeof value === "object" && !Array.isArray(value)) {
          flatten(value, path, paths)
        } else if (includeRe.test(path) && !excludeRe.test(path)) {
          paths.push([path, value])
        }
      }
      return paths
    }

    const getCookie = (name) => {
      const prefix = `${encodeURIComponent(name)}=`
      for (const part of document.cookie.split("; ")) {
        if (part.startsWith(prefix)) {
          return decodeURIComponent(part.slice(prefix.length))
        }
      }
      return ""
    }

    const setCookie = (name, value) => {
      const parts = [
        `${encodeURIComponent(name)}=${encodeURIComponent(value)}`,
        `Path=${cookiePath}`,
        `Max-Age=${maxAge}`,
        "SameSite=Lax",
      ]
      if (location.protocol === "https:") {
        parts.push("Secure")
      }
      document.cookie = parts.join("; ")
    }

    const applyCookieToSignals = () => {
      const raw = getCookie(cookieName)
      if (!raw) return
      let data
      try {
        data = JSON.parse(raw)
      } catch {
        return
      }
      if (data === null || typeof data !== "object" || Array.isArray(data)) {
        return
      }
      const paths = flatten(data)
      if (paths.length) {
        mergePaths(paths)
      }
    }

    applyCookieToSignals()

    let skipWrite = true
    let lastMarshalled = ""
    const stopEffect = effect(() => {
      const marshalled = JSON.stringify(filtered(signalFilter, root))
      if (skipWrite) {
        skipWrite = false
        lastMarshalled = marshalled
        return
      }
      if (marshalled === lastMarshalled) {
        return
      }
      lastMarshalled = marshalled
      setCookie(cookieName, marshalled)
    })

    queueMicrotask(() => {
      applyCookieToSignals()
    })

    return () => {
      stopEffect()
    }
  },
})
