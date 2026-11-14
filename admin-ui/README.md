# NoCloud Admin UI

Vue 3 + Vuetify 3 + Vite. Served under `/admin` by `apiserver_web`.

Requires Node.js 22.12+ (or 20.19+).

## Project setup
```
yarn install
```

### Dev server with hot reload
```
yarn dev
```
The dev server runs on http://localhost:8080/admin/. Runtime settings (colors,
theme, logo) come from `public/js/config.js`, which each installation mounts.

### Production build
```
yarn build
```
Output goes to `dist/`.

### Lint
```
yarn lint
```
