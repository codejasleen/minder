-- name: CreateAcceptedRisk :one
INSERT INTO accepted_risks (
    project_id,
    entity_id,
    rule_type_id,
    expires_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteExpiredAcceptedRisk :exec
DELETE FROM accepted_risks
WHERE project_id = $1
  AND entity_id = $2
  AND rule_type_id = $3
  AND expires_at <= NOW();

-- name: ListAcceptedRisksByProjectID :many
SELECT *
FROM accepted_risks
WHERE project_id = $1
  AND expires_at > NOW()
ORDER BY created_at ASC;

-- name: DeleteAcceptedRisk :exec
DELETE FROM accepted_risks
WHERE id = $1 AND project_id = $2;

-- name: GetActiveAcceptedRisk :one
SELECT *
FROM accepted_risks
WHERE project_id = $1
  AND entity_id = $2
  AND rule_type_id = $3
  AND expires_at > NOW();