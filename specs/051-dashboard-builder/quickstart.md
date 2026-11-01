# Quickstart: Dashboard Builder

1. Start PostgreSQL, load the demo data (it includes "Lubricants Europe overview", DB-1001, by Priya Raman, shared
   with Lubricants Europe, and the dataset "Event leads 2026", DS-1001) and run the API and the web app:

   ```sh
   cd backend && go run ./cmd/seed -reset && go run ./cmd/api
   cd frontend && pnpm run dev
   ```

2. Sign in as **Priya Raman** (Planner, Lubricants Europe). Open **Dashboards**: Built-in shows the Executive and
   Regional dashboards; My dashboards shows "Lubricants Europe overview". Open it, point at a bar, choose **Show as
   table** and **Download as PNG**; set the date range to July–December 2026 and use Back.
3. Choose **Dashboard builder**. Click **Bar**; in **Data** choose Activities, Group by Activity type, Sum of Planned
   cost. Drag the widget, then move it with the arrow keys and resize it with Shift + arrows. Press Ctrl+Z / ⌘Z.
   Name the dashboard and **Save**. Choose **Sharing** and tick Lubricants Europe.
4. Open **Datasets** › **Upload data** and pick `frontend/e2e/fixtures/event-leads.xlsx`; set "Cost" to Money in EUR,
   save, and chart it from the builder's source list (Uploaded data).
5. Sign in as **Sara Novak** (Viewer, both markets): the dashboard is under Shared with me with both plans' figures;
   the dataset widget shows "You don't have access to this data" until Priya shares the dataset. Sara has no
   Dashboard builder item.
6. Sign in as **Mei Lin** (Mobility China): Priya's dashboard is not listed, and `/dashboards/DB-1001` says
   "Dashboard not found".

Checks: `go test ./internal/customdashboard/... ./internal/dataset/... ./internal/widgetdata/... ./internal/app/...`
(`TEST_DATABASE_URL` for the scope-isolation tests); `pnpm vitest run src/features/custom-dashboards`;
`pnpm run test:e2e -- dashboard-builder custom-dashboards`.
