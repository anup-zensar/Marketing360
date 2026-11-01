-- name: ListActivitiesApprovedAt :many
SELECT subject_id AS activity_id, max(ended_at)::timestamptz AS approved_at
FROM approval_rounds
WHERE subject_kind = 'activity' AND outcome = 'approved' AND ended_at IS NOT NULL
    AND subject_id = ANY (sqlc.arg(activity_ids)::text[])
GROUP BY subject_id
ORDER BY subject_id;

-- name: ListPurchaseRequestsWaitingForGsap :many
SELECT p.id, p.name, p.vendor, p.plan_id, p.gsap_pr_number, o.name AS owner_name, r.type AS routing, r.routed_at,
    l.line_item_id, l.amount_minor, l.currency
FROM purchase_requests p
JOIN people o ON o.id = p.owner_id
JOIN LATERAL (SELECT g.type, g.routed_at FROM purchase_request_routings g
    WHERE g.purchase_request_id = p.id AND g.type IN ('original', 'modification')
    ORDER BY g.routed_at DESC, length(g.id) DESC, g.id COLLATE "C" DESC LIMIT 1) r ON true
JOIN purchase_request_lines l ON l.purchase_request_id = p.id
WHERE p.status = 'approved'
    AND EXISTS (SELECT 1 FROM purchase_request_routings g WHERE g.purchase_request_id = p.id)
    AND EXISTS (SELECT 1 FROM purchase_request_lines w WHERE w.purchase_request_id = p.id
        AND (w.po_line = '' OR w.po_value_minor <> w.amount_minor OR w.po_currency <> w.currency))
    AND EXISTS (SELECT 1 FROM purchase_request_lines m WHERE m.purchase_request_id = p.id
        AND m.line_item_id = ANY (sqlc.arg(line_item_ids)::text[]))
ORDER BY length(p.id), p.id COLLATE "C", l.position;

-- name: ListPurchaseOrderTransactions :many
SELECT t.id, t.type, t.line_item_id::text AS line_item_id, t.po_number, t.line_number, t.posting_date, t.amount_minor,
    t.currency, t.vendor_name, coalesce(pr.id, '')::text AS purchase_request_id,
    coalesce(pr.gsap_pr_number, '')::text AS gsap_pr_number
FROM financial_transactions t
LEFT JOIN LATERAL (SELECT p.id, p.gsap_pr_number FROM purchase_request_lines l
    JOIN purchase_requests p ON p.id = l.purchase_request_id
    WHERE l.po_number = t.po_number AND l.line_item_id = t.line_item_id
    ORDER BY length(p.id), p.id COLLATE "C" LIMIT 1) pr ON true
WHERE t.type IN ('purchaseOrder', 'invoice', 'payment') AND t.po_number <> ''
    AND t.line_item_id = ANY (sqlc.arg(line_item_ids)::text[])
ORDER BY length(t.id), t.id COLLATE "C";
