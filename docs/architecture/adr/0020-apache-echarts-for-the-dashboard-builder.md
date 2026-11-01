# ADR-0020: Apache ECharts for the dashboard builder's charts

**Status:** Accepted · **Date:** 2026-11-01

## Context

Feature 051 (constitution 4.9.0, D-371) lets people build their own dashboards from twelve chart types — bar
(vertical and horizontal; grouped, stacked and 100 %), line, area, combined bar and line on two axes, pie, donut,
scatter, funnel, gauge, heatmap, treemap and radar — beside figure tiles, tables and text notes. ADR-0017 drew
036's two dashboards as hand-written SVG bars with no chart library. Writing twelve chart types by hand —
axes, scales, stacking, legends, tooltips, zoom, layout of labels — would be a charting library of our own, with
its own bugs and no maintainer but us.

Constitution F1 requires one charting approach chosen by ADR that records accessibility conformance,
maintenance activity, licence and bundle size; F6 a data table or text summary for every chart; F7 and
ADR-0018 a content security policy without evaluated code or inline style attributes; F8 heavy libraries
loaded only with the screens that need them; VII (release age) versions published at least 24 hours before
they are installed. The product owner named Apache ECharts (2026-11-01).

## Decision

- **Apache ECharts 6.1.0** (published 2026-05-19, so older than pnpm's `minimumReleaseAge`) is the chart library
  of the dashboard builder and the custom-dashboard viewer. Its dependencies are `zrender` 6.1.0 (BSD-3-Clause,
  same team) and `tslib` 2.3.0 (0BSD); none runs an install script (`prepare` runs only when the package itself
  is built from its repository). Added with `pnpm add echarts@6.1.0`, pinned exactly.
- **Imported modularly** from `echarts/core`, registering only what the builder uses: the SVG renderer; the
  components grid, tooltip, legend, title, dataset, aria, visualMap (heatmap), radar, dataZoom (zoom slider) and
  markLine (reference line); the label-layout feature; and the charts bar, line, pie, scatter, funnel, gauge,
  heatmap, treemap and radar (area, combo and donut are options of line, bar + line and pie). Never `import
  * as echarts from 'echarts'`, which brings everything.
- **SVG renderer, not canvas.** At most 500 groups per widget (D-373) is far below the point counts where canvas
  wins; SVG stays sharp at any zoom and in a PNG drawn at twice the size, and zrender's SVG painter writes
  styles through the CSSOM (`element.style.cssText`), which the content security policy allows.
- **Our own small component** (`EChart`, about 80 lines) creates the chart with `echarts.init(element, null,
  { renderer: 'svg', locale: 'EN' })`, sets the option, resizes with a `ResizeObserver`, re-renders when the
  theme changes and disposes of the chart on unmount. No `echarts-for-react` or other wrapper.
- **Options are built by pure functions** (`buildOption(widget, data, theme)`), one per chart type, from the
  API's figures only (constitution IV): they are unit-tested; ECharts itself is not.
- **Accessibility**: the `aria` option is on, with a description written by us (title, type, the dimension and
  measures, the number of points) instead of ECharts' generated one; `aria.decal.show` adds patterns to bars,
  areas and slices so colour is never the only carrier (WCAG 1.4.1); every chart has **Show as table**, a real
  `<table>` of the same rows (F6); the legend names every series in words; tooltips repeat what the table holds.
  ECharts makes no WCAG conformance claim of its own, which is why the table view is mandatory, not optional.
- **Theming from tokens**: colours, fonts and lines are read from the design tokens (`getComputedStyle` on the
  chart's element) when the option is built — the four chart palettes of design system §4.7 and Text Strong,
  Text Subtle and Border Subtle for text, axes and gridlines — and rebuilt when the theme changes. No ECharts
  theme is registered, so there is one source of colour: `tokens.css`.
- **Content security policy (ADR-0018)**: ECharts' HTML tooltip writes HTML strings with inline `style`
  attributes into `innerHTML`, which `style-src 'self'` refuses. Every tooltip therefore uses a `formatter`
  that returns DOM nodes built with `textContent` and CSS-module classes, which ECharts appends as they are
  (`TooltipHTMLContent.setContent` accepts elements), and sets `confine: true`, so ECharts adds no tooltip
  arrow (another HTML string); text from data is never parsed as HTML. The only
  evaluated code in ECharts (`new Function` when a map's GeoJSON is given as a string) is in the geo component,
  which is not imported. The toolbox (its data view writes HTML) is not used. The end-to-end tests fail on any
  policy violation (`e2e/support/no-csp-violations.ts`).
- **PNG download**: with the SVG renderer `getDataURL()` gives SVG, so the browser draws that SVG (a `data:`
  image, allowed by `img-src 'self' data:`) on a canvas at twice the size with the surface colour behind it and
  saves `canvas.toBlob('image/png')`. Nothing leaves the machine.
- **Loaded with its pages** (F8): only `features/custom-dashboards` imports ECharts, and its pages are lazy
  routes, so no other page downloads it.
- **036's dashboards keep their SVG charts** (ADR-0017) until a later change moves them; the builder's grid is
  hand-built (CSS grid with pointer and keyboard moves), not a grid library.

## Bundle size

Measured on 2026-11-01 with esbuild 0.25 (minified ES modules, gzip -9) from `echarts@6.1.0`:

| Bundle | Minified | Gzipped |
|---|---|---|
| `echarts/core` + SVG renderer only | 315 KB | 108 KB |
| + grid, tooltip, legend, title, dataset, aria components and label layout | 529 KB | 181 KB |
| the same with the canvas renderer instead (for comparison) | 543 KB | 186 KB |
| + bar chart | 555 KB | 190 KB |
| + line chart | 559 KB | 192 KB |
| + pie chart | 545 KB | 187 KB |
| + scatter chart | 543 KB | 186 KB |
| + funnel chart | 537 KB | 184 KB |
| + gauge chart | 539 KB | 185 KB |
| + heatmap chart and visualMap | 574 KB | 195 KB |
| + treemap chart | 569 KB | 195 KB |
| + radar chart and radar component | 541 KB | 185 KB |
| **Everything the builder registers** (all of the above, dataZoom, markLine) | **759 KB** | **256 KB** |
| The whole library (`import * from 'echarts'`), not used | 1,141 KB | 382 KB |

The custom-dashboards chunk therefore adds about 256 KB gzipped to the pages that draw charts and nothing to
any other page.

## Alternatives considered

- **More hand-written SVG (ADR-0017's approach)**: fine for one grouped-bar chart; twelve types with stacking,
  dual axes, zoom, treemap layout and radar geometry would be a library of our own. Rejected.
- **Recharts / Nivo / Victory** (React + D3): good React fit, but none covers funnel, gauge, heatmap, treemap and
  radar together with one consistent option model; several pull in many D3 modules and SVG `style` attributes.
  Rejected.
- **Chart.js**: canvas only (no SVG to read in tests or to draw sharply in a PNG at any size), and treemap,
  funnel and heatmap need third-party plugins. Rejected.
- **Highcharts**: commercial licence. Rejected.
- **Vega-Lite**: compiles specifications at run time with expression evaluation that needs `unsafe-eval` unless
  configured with its interpreter; large. Rejected.
- **`echarts-for-react`**: a wrapper of about 100 lines that imports all of ECharts by default and lags ECharts
  releases. Rejected for our own component.
- **Canvas renderer**: same size; blurry when scaled, nothing to inspect in a browser test. Rejected.

## Consequences

- ECharts is an architecture-shaping dependency: upgrades follow pnpm's release-age rule and are checked with
  the builder's end-to-end tests (rendering, CSP, axe).
- Tests check `buildOption` (the option for each type and display choice), the data states and the table view,
  not ECharts' drawing; component tests mock `EChart` where a real SVG adds nothing.
- The ECharts licence (Apache-2.0) and its NOTICE are recorded in the pull request that adds it; zrender's
  BSD-3-Clause and tslib's 0BSD with it.
- Two chart approaches live side by side (036's SVG, the builder's ECharts) until 036's charts are moved; the
  constitution's F1 row says so (4.9.0).
