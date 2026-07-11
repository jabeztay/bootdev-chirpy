-- name: DeleteChirpById :exec
DELETE FROM chirps
WHERE id = $1;
